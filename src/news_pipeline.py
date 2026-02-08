import json
import math
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Dict, Iterable, List, Optional, Set, Tuple

import feedparser
import requests
from dateutil import parser

DATA_DIR = Path(__file__).resolve().parents[1] / "data"
SOURCES_PATH = DATA_DIR / "sources.json"
WEIGHTS_PATH = DATA_DIR / "weights.json"

TOTAL_DAILY_SLOTS = 10
MAX_SLOTS_PER_SOURCE = 3
REQUIRED_RESEARCH_AI_AUTO = 2
TARGET_CATEGORIES = {"ai", "auto", "games", "politics"}
ALGORITHM_VERSION = "v2-impact-weighted"

CATEGORIES = {
    "ai": ["ai", "artificial intelligence", "machine learning", "llm", "robot", "deep learning"],
    "auto": ["auto", "car", "vehicle", "ev", "electric vehicle", "autonomous", "battery"],
    "games": ["game", "gaming", "esports", "console"],
    "politics": ["election", "government", "policy", "minister", "president", "parliament", "congress"],
}

RESEARCH_KEYWORDS = [
    "study",
    "research",
    "paper",
    "journal",
    "university",
    "conference",
    "arxiv",
    "nature",
    "science",
]


@dataclass
class Source:
    id: str
    name: str
    country: str
    rss: str
    base_authority: float
    topics: List[str]


@dataclass
class Article:
    source_id: str
    source_name: str
    title: str
    link: str
    published_at: datetime
    summary: str
    categories: List[str]
    is_research: bool


@dataclass
class DigestMeta:
    generated_at: datetime
    research_target: int
    research_selected: int
    notes: List[str]


def load_sources() -> List[Source]:
    with open(SOURCES_PATH, "r", encoding="utf-8") as handle:
        raw = json.load(handle)
    return [Source(**item) for item in raw]


def fetch_feed(url: str) -> feedparser.FeedParserDict:
    response = requests.get(url, timeout=15)
    response.raise_for_status()
    return feedparser.parse(response.content)


def normalize_text(text: str) -> str:
    return text.lower()


def detect_categories(title: str, summary: str) -> Tuple[List[str], bool]:
    text = normalize_text(f"{title} {summary}")
    categories = [
        category
        for category, keywords in CATEGORIES.items()
        if any(keyword in text for keyword in keywords)
    ]
    is_research = any(keyword in text for keyword in RESEARCH_KEYWORDS)
    return categories, is_research


def parse_datetime(value: Optional[str]) -> Optional[datetime]:
    if not value:
        return None
    try:
        parsed = parser.parse(value)
    except (parser.ParserError, TypeError, ValueError):
        return None
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    return parsed.astimezone(timezone.utc)


def is_trustworthy_article(source_ids: Set[str], source_id: str, title: str, link: str) -> bool:
    if source_id not in source_ids:
        return False
    if len(title.strip()) < 8:
        return False
    if not link.startswith("http://") and not link.startswith("https://"):
        return False
    return True


def load_articles(sources: Iterable[Source]) -> List[Article]:
    source_list = list(sources)
    allowed_ids = {source.id for source in source_list}
    articles: List[Article] = []
    for source in source_list:
        try:
            feed = fetch_feed(source.rss)
        except requests.RequestException:
            continue
        for entry in feed.entries:
            title = entry.get("title", "").strip()
            link = entry.get("link", "").strip()
            if not is_trustworthy_article(allowed_ids, source.id, title, link):
                continue
            summary = entry.get("summary", "") or entry.get("description", "") or ""
            published = parse_datetime(entry.get("published") or entry.get("updated"))
            if not published:
                continue
            categories, is_research = detect_categories(title, summary)
            categories = [c for c in categories if c in TARGET_CATEGORIES]
            if not categories:
                continue
            articles.append(
                Article(
                    source_id=source.id,
                    source_name=source.name,
                    title=title,
                    link=link,
                    published_at=published,
                    summary=summary,
                    categories=categories,
                    is_research=is_research,
                )
            )
    return articles


def compute_weekly_scores(sources: Iterable[Source], articles: Iterable[Article]) -> Tuple[Dict[str, float], Dict[str, Dict[str, float]]]:
    now = datetime.now(timezone.utc)
    week_ago = now - timedelta(days=7)
    source_map = {source.id: source for source in sources}
    weekly = [a for a in articles if a.published_at >= week_ago and a.source_id in source_map]

    counts = {sid: 0 for sid in source_map}
    research_counts = {sid: 0 for sid in source_map}
    topic_sets: Dict[str, Set[str]] = {sid: set() for sid in source_map}

    for article in weekly:
        counts[article.source_id] += 1
        if article.is_research:
            research_counts[article.source_id] += 1
        topic_sets[article.source_id].update(article.categories)

    scores: Dict[str, float] = {}
    metrics: Dict[str, Dict[str, float]] = {}
    for sid, source in source_map.items():
        volume = counts[sid]
        volume_impact = math.log1p(volume)
        research_ratio = (research_counts[sid] / volume) if volume > 0 else 0.0
        topic_coverage = len(topic_sets[sid] & TARGET_CATEGORIES) / float(len(TARGET_CATEGORIES))

        impact_factor_like = 1.0 + (0.45 * volume_impact) + (0.35 * research_ratio) + (0.20 * topic_coverage)
        score = round(source.base_authority * impact_factor_like, 4)
        scores[sid] = score
        metrics[sid] = {
            "base_authority": round(source.base_authority, 4),
            "weekly_volume": float(volume),
            "volume_impact": round(volume_impact, 4),
            "research_ratio": round(research_ratio, 4),
            "topic_coverage": round(topic_coverage, 4),
            "impact_factor_like": round(impact_factor_like, 4),
            "score": score,
        }
    return scores, metrics


def load_or_update_scores(sources: Iterable[Source], articles: Iterable[Article]) -> Tuple[Dict[str, float], Dict[str, Dict[str, float]]]:
    source_ids = {source.id for source in sources}
    if WEIGHTS_PATH.exists():
        with open(WEIGHTS_PATH, "r", encoding="utf-8") as handle:
            payload = json.load(handle)
        last_updated = parse_datetime(payload.get("updated_at"))
        same_algo = payload.get("algorithm_version") == ALGORITHM_VERSION
        saved_scores = payload.get("scores", {})
        if (
            same_algo
            and last_updated
            and datetime.now(timezone.utc) - last_updated < timedelta(days=7)
            and all(source_id in saved_scores for source_id in source_ids)
        ):
            return saved_scores, payload.get("metrics", {})

    scores, metrics = compute_weekly_scores(sources, articles)
    payload = {
        "updated_at": datetime.now(timezone.utc).isoformat(),
        "algorithm_version": ALGORITHM_VERSION,
        "scores": scores,
        "metrics": metrics,
    }
    with open(WEIGHTS_PATH, "w", encoding="utf-8") as handle:
        json.dump(payload, handle, ensure_ascii=False, indent=2)
    return scores, metrics


def allocate_slots(scores: Dict[str, float], total_slots: int = TOTAL_DAILY_SLOTS) -> Dict[str, int]:
    total_score = sum(scores.values())
    if total_score <= 0:
        return {source_id: 0 for source_id in scores}

    raw = {source_id: total_slots * score / total_score for source_id, score in scores.items()}
    slots = {source_id: min(MAX_SLOTS_PER_SOURCE, int(math.floor(value))) for source_id, value in raw.items()}

    remainder = total_slots - sum(slots.values())
    ranked = sorted(raw.items(), key=lambda item: item[1] - math.floor(item[1]), reverse=True)
    while remainder > 0:
        moved = False
        for source_id, _ in ranked:
            if slots[source_id] >= MAX_SLOTS_PER_SOURCE:
                continue
            slots[source_id] += 1
            remainder -= 1
            moved = True
            if remainder <= 0:
                break
        if not moved:
            break
    return slots


def select_articles(
    articles: Iterable[Article],
    slots: Dict[str, int],
    scores: Dict[str, float],
    require_research: int = REQUIRED_RESEARCH_AI_AUTO,
) -> Tuple[List[Article], DigestMeta]:
    article_list = list(articles)
    articles_by_source: Dict[str, List[Article]] = {}
    for article in article_list:
        articles_by_source.setdefault(article.source_id, []).append(article)
    for grouped in articles_by_source.values():
        grouped.sort(key=lambda item: item.published_at, reverse=True)

    selected: List[Article] = []
    selected_links: Set[str] = set()

    research_pool = [
        article
        for article in article_list
        if article.is_research and ("ai" in article.categories or "auto" in article.categories)
    ]
    research_pool.sort(
        key=lambda item: (scores.get(item.source_id, 0.0), item.published_at),
        reverse=True,
    )

    for article in research_pool:
        if len(selected) >= require_research:
            break
        if slots.get(article.source_id, 0) <= 0:
            continue
        if article.link in selected_links:
            continue
        selected.append(article)
        selected_links.add(article.link)
        slots[article.source_id] -= 1

    source_rank = sorted(slots.keys(), key=lambda sid: scores.get(sid, 0.0), reverse=True)
    for source_id in source_rank:
        remaining = slots.get(source_id, 0)
        if remaining <= 0:
            continue
        for article in articles_by_source.get(source_id, []):
            if remaining <= 0:
                break
            if article.link in selected_links:
                continue
            selected.append(article)
            selected_links.add(article.link)
            remaining -= 1
        slots[source_id] = remaining

    selected.sort(key=lambda item: item.published_at, reverse=True)
    selected = selected[:TOTAL_DAILY_SLOTS]

    notes: List[str] = []
    if len(selected) < TOTAL_DAILY_SLOTS:
        notes.append(f"今日满足条件新闻不足，仅输出 {len(selected)} 篇。")
    if len([a for a in selected if a.is_research and ('ai' in a.categories or 'auto' in a.categories)]) < require_research:
        notes.append("今日 AI/汽车前沿研究类新闻不足 2 篇，已输出可获取的全部研究类内容。")

    meta = DigestMeta(
        generated_at=datetime.now(timezone.utc),
        research_target=require_research,
        research_selected=len([a for a in selected if a.is_research and ('ai' in a.categories or 'auto' in a.categories)]),
        notes=notes,
    )
    return selected, meta


def build_daily_digest() -> Tuple[List[Article], Dict[str, float], Dict[str, int], Dict[str, Dict[str, float]], DigestMeta]:
    sources = load_sources()
    articles = load_articles(sources)
    scores, metrics = load_or_update_scores(sources, articles)
    slots = allocate_slots(scores)

    today = datetime.now(timezone.utc).date()
    today_articles = [
        article
        for article in articles
        if article.published_at.date() == today and any(c in TARGET_CATEGORIES for c in article.categories)
    ]
    digest, meta = select_articles(today_articles, slots, scores)
    return digest, scores, slots, metrics, meta
