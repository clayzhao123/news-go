import json
from dataclasses import asdict
from pathlib import Path

from src.news_pipeline import WEIGHTS_PATH, build_daily_digest

DIGEST_PATH = Path(__file__).resolve().parents[1] / "data" / "daily_digest.json"


def main() -> None:
    digest, scores, slots, metrics, meta = build_daily_digest()
    payload = {
        "generated_at": meta.generated_at.isoformat(),
        "research_target": meta.research_target,
        "research_selected": meta.research_selected,
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
    DIGEST_PATH.parent.mkdir(parents=True, exist_ok=True)
    DIGEST_PATH.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"daily digest written: {DIGEST_PATH}")


if __name__ == "__main__":
    main()
