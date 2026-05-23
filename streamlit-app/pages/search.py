import streamlit as st
import sys, os

sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))
from theme import inject_theme, sidebar_logo, page_header

st.set_page_config(
    page_title="Search · Know-Blocks",
    page_icon="🔍",
    layout="wide",
    initial_sidebar_state="expanded",
)

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
    "Know-Blocks · Search",
    "Chunk",
    "Search",
    "// full-text search across all chunks in the document store",
)

if not st.session_state.documents:
    st.markdown(
        """
    <div class="empty-state">
        <div class="empty-icon">🔍</div>
        <div class="empty-title">No documents in the store yet</div>
        <div class="empty-sub">Deploy a document first to start searching chunks</div>
    </div>
    """,
        unsafe_allow_html=True,
    )
    col, _ = st.columns([1, 3])
    with col:
        if st.button("⬆  Go to Deploy", use_container_width=True):
            st.switch_page("pages/deploy.py")
else:
    st.markdown(
        """
    <div class="kb-info">
        ▸ Search across all chunks in your document store · Select chunks to add to a collection
    </div>
    """,
        unsafe_allow_html=True,
    )

    st.markdown(
        """
    <div class="empty-state">
        <div class="empty-icon">🔍</div>
        <div class="empty-title">Search coming soon</div>
        <div class="empty-sub">Full-text search across chunks · semantic search within collections</div>
    </div>
    """,
        unsafe_allow_html=True,
    )
