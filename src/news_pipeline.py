import json
import math
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Dict, Iterable, List, Optional, Tuple

import feedparser
import requests
from dateutil import parser

DATA_DIR = Path(__file__).resolve().parents[1] / "data"
SOURCES_PATH = DATA_DIR / "sources.json"
WEIGHTS_PATH = DATA_DIR / "weights.json"

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


def load_articles(sources: Iterable[Source]) -> List[Article]:
    articles: List[Article] = []
    for source in sources:
        feed = fetch_feed(source.rss)
        for entry in feed.entries:
            title = entry.get("title", "").strip()
            link = entry.get("link", "").strip()
            summary = entry.get("summary", "") or entry.get("description", "") or ""
            published = parse_datetime(entry.get("published") or entry.get("updated"))
            if not published:
                continue
            categories, is_research = detect_categories(title, summary)
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


def compute_weekly_scores(sources: Iterable[Source], articles: Iterable[Article]) -> Dict[str, float]:
    now = datetime.now(timezone.utc)
    week_ago = now - timedelta(days=7)
    counts = {source.id: 0 for source in sources}
    for article in articles:
        if article.published_at >= week_ago:
            counts[article.source_id] += 1
    scores: Dict[str, float] = {}
    for source in sources:
        volume = counts.get(source.id, 0)
        impact = math.log1p(volume)
        scores[source.id] = round(source.base_authority * (1.0 + impact), 4)
    return scores


def load_or_update_scores(sources: Iterable[Source], articles: Iterable[Article]) -> Dict[str, float]:
    if WEIGHTS_PATH.exists():
        with open(WEIGHTS_PATH, "r", encoding="utf-8") as handle:
            payload = json.load(handle)
        last_updated = parse_datetime(payload.get("updated_at"))
        if last_updated and datetime.now(timezone.utc) - last_updated < timedelta(days=7):
            return payload.get("scores", {})
    scores = compute_weekly_scores(sources, articles)
    payload = {
        "updated_at": datetime.now(timezone.utc).isoformat(),
        "scores": scores,
    }
    with open(WEIGHTS_PATH, "w", encoding="utf-8") as handle:
        json.dump(payload, handle, ensure_ascii=False, indent=2)
    return scores


def allocate_slots(scores: Dict[str, float], total_slots: int = 10) -> Dict[str, int]:
    total_score = sum(scores.values())
    if total_score == 0:
        return {source_id: 0 for source_id in scores}
    raw = {source_id: total_slots * score / total_score for source_id, score in scores.items()}
    slots = {source_id: int(math.floor(value)) for source_id, value in raw.items()}
    remainder = total_slots - sum(slots.values())
    for source_id, _ in sorted(raw.items(), key=lambda item: item[1] - math.floor(item[1]), reverse=True):
        if remainder <= 0:
            break
        slots[source_id] += 1
        remainder -= 1
    return slots


def select_articles(
    articles: Iterable[Article],
    slots: Dict[str, int],
    require_research: int = 2,
) -> List[Article]:
    articles_by_source: Dict[str, List[Article]] = {}
    for article in articles:
        articles_by_source.setdefault(article.source_id, []).append(article)
    for article_list in articles_by_source.values():
        article_list.sort(key=lambda item: item.published_at, reverse=True)

    selected: List[Article] = []
    research_pool = [
        article
        for article in articles
        if article.is_research and ("ai" in article.categories or "auto" in article.categories)
    ]
    research_pool.sort(key=lambda item: item.published_at, reverse=True)
    for article in research_pool:
        if len(selected) >= require_research:
            break
        if slots.get(article.source_id, 0) <= 0:
            continue
        selected.append(article)
        slots[article.source_id] -= 1

    for source_id, remaining in slots.items():
        if remaining <= 0:
            continue
        for article in articles_by_source.get(source_id, []):
            if article in selected:
                continue
            selected.append(article)
            remaining -= 1
            if remaining <= 0:
                break

    selected.sort(key=lambda item: item.published_at, reverse=True)
    return selected[:10]


def build_daily_digest() -> Tuple[List[Article], Dict[str, float], Dict[str, int]]:
    sources = load_sources()
    articles = load_articles(sources)
    scores = load_or_update_scores(sources, articles)
    slots = allocate_slots(scores)
    today = datetime.now(timezone.utc).date()
    today_articles = [article for article in articles if article.published_at.date() == today]
    digest = select_articles(today_articles, slots)
    return digest, scores, slots
