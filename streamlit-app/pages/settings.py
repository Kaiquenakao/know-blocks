import streamlit as st
import sys, os

sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))
from theme import inject_theme, sidebar_logo, page_header

st.set_page_config(page_title="Settings · Know-Blocks", page_icon="⚙", layout="wide")
inject_theme()

for key, val in [("documents", []), ("total_chunks", 0), ("collections", [])]:
    if key not in st.session_state:
        st.session_state[key] = val

with st.sidebar:
    sidebar_logo()
    st.markdown('<div class="nav-section">Workspace</div>', unsafe_allow_html=True)
    st.page_link("app.py", label="🏠  Document Store")
    st.page_link("pages/deploy.py", label="⬆  Deploy")
    st.page_link("pages/search.py", label="🔍  Search Chunks")
    st.markdown('<div class="nav-section">Manage</div>', unsafe_allow_html=True)
    st.page_link("pages/collections.py", label="🗂  Collections")
    st.page_link("pages/settings.py", label="⚙  Settings")

page_header(
    "Know-Blocks · Config", "Platform", "Settings", "// configure defaults and API keys"
)

st.markdown(
    """
<div class="empty-state">
    <div class="empty-icon">⚙</div>
    <div class="empty-title">Settings — coming soon</div>
    <div class="empty-sub">Default embedding model · API keys · storage config · chunking presets</div>
</div>
""",
    unsafe_allow_html=True,
)
