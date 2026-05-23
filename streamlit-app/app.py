import streamlit as st
import sys, os
from dotenv import load_dotenv
load_dotenv()
sys.path.insert(0, os.path.dirname(__file__))
from theme import inject_theme, sidebar_logo, page_header

st.set_page_config(
    page_title="Know-Blocks",
    page_icon="🧱",
    layout="wide",
    initial_sidebar_state="expanded",
)

inject_theme()

# ── Session state defaults ────────────────────────────────────────────────────
if "documents" not in st.session_state:
    st.session_state.documents = []
if "total_chunks" not in st.session_state:
    st.session_state.total_chunks = 0
if "collections" not in st.session_state:
    st.session_state.collections = []

# ── Sidebar ───────────────────────────────────────────────────────────────────
with st.sidebar:
    sidebar_logo()
    st.markdown('<div class="nav-section">Workspace</div>', unsafe_allow_html=True)
    st.page_link("app.py",            label="🏠  Document Store",  )
    st.page_link("pages/deploy.py", label="⬆  Deploy")
    st.page_link("pages/search.py", label="🔍  Search Chunks")
    st.markdown('<div class="nav-section">Manage</div>', unsafe_allow_html=True)
    st.page_link("pages/collections.py", label="🗂  Collections")
    st.page_link("pages/settings.py",    label="⚙  Settings")

    # Platform stats footer
    st.markdown("<br>" * 4, unsafe_allow_html=True)
    docs = len(st.session_state.documents)
    chunks = st.session_state.total_chunks
    cols_count = len(st.session_state.collections)
    st.markdown(f"""
    <div style="border-top:1px solid #3d1515; padding:1rem 0.5rem; margin-top:auto;">
        <div style="font-family:'DM Mono',monospace; font-size:0.58rem; color:#4a1a1a;
                    letter-spacing:2px; text-transform:uppercase; margin-bottom:0.8rem;">
            Platform
        </div>
        <div style="display:flex; justify-content:space-between; margin-bottom:0.35rem;">
            <span style="font-size:0.78rem; color:#6b3030;">Documents</span>
            <span style="font-family:'DM Mono',monospace; font-size:0.78rem; color:#DCC3AA;">{docs}</span>
        </div>
        <div style="display:flex; justify-content:space-between; margin-bottom:0.35rem;">
            <span style="font-size:0.78rem; color:#6b3030;">Chunks</span>
            <span style="font-family:'DM Mono',monospace; font-size:0.78rem; color:#DCC3AA;">{chunks}</span>
        </div>
        <div style="display:flex; justify-content:space-between;">
            <span style="font-size:0.78rem; color:#6b3030;">Collections</span>
            <span style="font-family:'DM Mono',monospace; font-size:0.78rem; color:#DCC3AA;">{cols_count}</span>
        </div>
    </div>
    """, unsafe_allow_html=True)

# ── Main content ──────────────────────────────────────────────────────────────
page_header(
    "Know-Blocks · Home",
    "Document",
    "Store",
    "// all deployed documents and their chunk status"
)

# Stats row
docs = st.session_state.documents
c1, c2, c3, c4 = st.columns(4)
for col, val, label in [
    (c1, len(docs), "Documents"),
    (c2, st.session_state.total_chunks, "Total Chunks"),
    (c3, len(st.session_state.collections), "Collections"),
    (c4, sum(d.get("embeddings", 0) for d in docs), "Embeddings"),
]:
    col.markdown(f"""
    <div class="stat-card">
        <div class="stat-value">{val}</div>
        <div class="stat-label">{label}</div>
    </div>
    """, unsafe_allow_html=True)

st.markdown("<div style='margin-bottom:2rem'></div>", unsafe_allow_html=True)

# ── Document list ─────────────────────────────────────────────────────────────
st.markdown('<div class="kb-card-title">Deployed Documents</div>', unsafe_allow_html=True)

if not docs:
    st.markdown("""
    <div class="empty-state">
        <div class="empty-icon">📂</div>
        <div class="empty-title">No documents yet</div>
        <div class="empty-sub">Go to Deploy to upload your first document</div>
    </div>
    """, unsafe_allow_html=True)

    col_btn, _ = st.columns([1, 3])
    with col_btn:
        if st.button("⬆  Deploy first document", use_container_width=True):
            st.switch_page("pages/deploy.py")
else:
    # Search / filter bar
    search = st.text_input(
        "Filter documents",
        placeholder="Search by name, model or strategy...",
        label_visibility="collapsed",
    )

    st.markdown("<div style='margin-bottom:0.5rem'></div>", unsafe_allow_html=True)

    filtered = [
        d for d in docs
        if not search or search.lower() in d["name"].lower()
        or search.lower() in d.get("model", "").lower()
    ]

    if not filtered:
        st.markdown("""
        <div class="empty-state">
            <div class="empty-icon">🔎</div>
            <div class="empty-title">No results</div>
            <div class="empty-sub">Try a different search term</div>
        </div>
        """, unsafe_allow_html=True)

    for doc in filtered:
        model_short = doc.get("model", "—").split("(")[0].strip()
        strategy    = doc.get("strategy", "Fixed Size")
        chunks      = doc.get("chunks", 0)
        ext         = doc["name"].split(".")[-1].upper()
        deployed_at = doc.get("deployed_at", "—")

        st.markdown(f"""
        <div class="doc-row">
            <div style="font-size:1.6rem; opacity:0.7;">📄</div>
            <div style="flex:1; min-width:0;">
                <div style="font-weight:600; color:#F1E2D1; font-size:0.9rem;
                            white-space:nowrap; overflow:hidden; text-overflow:ellipsis;">
                    {doc["name"]}
                </div>
                <div style="font-family:'DM Mono',monospace; font-size:0.62rem;
                            color:#6b3030; margin-top:0.2rem; letter-spacing:0.5px;">
                    {deployed_at}
                </div>
            </div>
            <div style="text-align:center; min-width:60px;">
                <div style="font-family:'DM Mono',monospace; font-size:1rem;
                            font-weight:500; color:#DCC3AA;">{chunks}</div>
                <div style="font-family:'DM Mono',monospace; font-size:0.58rem;
                            color:#4a1a1a; letter-spacing:1px; text-transform:uppercase;">chunks</div>
            </div>
            <div style="min-width:120px;">
                <span class="kb-tag">{ext}</span>
                <span class="kb-tag">{strategy[:5]}</span>
            </div>
            <div style="font-family:'DM Mono',monospace; font-size:0.7rem;
                        color:#810B38; min-width:140px; text-align:right;">
                {model_short}
            </div>
        </div>
        """, unsafe_allow_html=True)

    st.markdown("<div style='margin-top:1.5rem'></div>", unsafe_allow_html=True)
    if st.button("⬆  Deploy another document"):
        st.switch_page("pages/deploy.py")