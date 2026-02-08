import json
from dataclasses import asdict
from pathlib import Path

from src.news_pipeline import WEIGHTS_PATH, build_daily_digest

DIGEST_PATH = Path(__file__).resolve().parents[1] / "data" / "daily_digest.json"


def _load_existing() -> dict:
    if not DIGEST_PATH.exists():
        return {}
    try:
        return json.loads(DIGEST_PATH.read_text(encoding="utf-8"))
    except (json.JSONDecodeError, OSError):
        return {}


def main() -> None:
    digest, scores, slots, metrics, meta = build_daily_digest()
    payload = {
        "generated_at": meta.generated_at.isoformat(),
        "research_target": meta.research_target,
        "research_selected": meta.research_selected,
        "notes": list(meta.notes),
        "notes": meta.notes,
        "scores": scores,
        "slots": slots,
        "metrics": metrics,
        "weights_file": str(WEIGHTS_PATH),
        "items": [
            {
                **asdict(article),
                "published_at": article.published_at.isoformat(),
            }
            for article in digest
        ],
    }

    existing = _load_existing()
    existing_items = existing.get("items", []) if isinstance(existing, dict) else []
    if not payload["items"] and existing_items:
        payload["items"] = existing_items
        payload["notes"].append("今日抓取为空，已回退到最近一次成功摘要。")
        payload["fallback_from_previous"] = True
    else:
        payload["fallback_from_previous"] = False

    DIGEST_PATH.parent.mkdir(parents=True, exist_ok=True)
    DIGEST_PATH.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"daily digest written: {DIGEST_PATH}")
    print(f"items: {len(payload['items'])}, fallback_from_previous: {payload['fallback_from_previous']}")
    DIGEST_PATH.parent.mkdir(parents=True, exist_ok=True)
    DIGEST_PATH.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"daily digest written: {DIGEST_PATH}")


if __name__ == "__main__":
    main()
