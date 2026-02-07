from datetime import datetime

import streamlit as st

from src.news_pipeline import build_daily_digest

st.set_page_config(page_title="News Digest", layout="wide")

st.title("每日新闻摘要")

if st.button("刷新今日摘要") or "digest" not in st.session_state:
    with st.spinner("抓取新闻并计算权重中..."):
        digest, scores, slots = build_daily_digest()
        st.session_state["digest"] = digest
        st.session_state["scores"] = scores
        st.session_state["slots"] = slots

st.subheader("来源权重（每周更新）")
if "scores" in st.session_state:
    st.dataframe(
        [
            {
                "source_id": source_id,
                "score": score,
                "slots_today": st.session_state["slots"].get(source_id, 0),
            }
            for source_id, score in sorted(
                st.session_state["scores"].items(), key=lambda item: item[1], reverse=True
            )
        ],
        use_container_width=True,
    )

st.subheader("今日新闻（最多10篇）")
if "digest" in st.session_state:
    for article in st.session_state["digest"]:
        with st.container(border=True):
            st.markdown(f"### {article.title}")
            st.markdown(f"**来源**：{article.source_name}")
            st.markdown(
                f"**发布时间**：{article.published_at.strftime('%Y-%m-%d %H:%M')}"
            )
            st.markdown(f"**分类**：{', '.join(article.categories)}")
            st.markdown(f"**研究类**：{'是' if article.is_research else '否'}")
            st.markdown(f"[阅读原文]({article.link})")
            if article.summary:
                st.markdown(article.summary)

st.caption(f"更新时间：{datetime.now().strftime('%Y-%m-%d %H:%M')}")
