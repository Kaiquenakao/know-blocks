THEME_CSS = """
<style>
@import url('https://fonts.googleapis.com/css2?family=Playfair+Display:wght@400;600;700;900&family=DM+Mono:wght@300;400;500&family=DM+Sans:wght@300;400;500;600&display=swap');

html, body, [class*="css"] { font-family: 'DM Sans', sans-serif; }
#MainMenu, footer, header { visibility: hidden; }

.stApp {
    background-color: #F1E2D1;
    color: #541A1A;
}

/* ── Sidebar dark maroon ── */
[data-testid="stSidebar"] {
    background-color: #541A1A;
    border-right: none;
}
[data-testid="stSidebar"] * { font-family: 'DM Sans', sans-serif !important; }

.kb-logo {
    padding: 2rem 1.2rem 1.5rem;
    border-bottom: 1px solid #6b2a2a;
    margin-bottom: 1rem;
}
.kb-logo-text {
    font-family: 'Playfair Display', serif;
    font-size: 1.5rem;
    font-weight: 900;
    color: #F1E2D1;
    letter-spacing: -0.5px;
}
.kb-logo-text span { color: #DCC3AA; }
.kb-logo-sub {
    font-family: 'DM Mono', monospace;
    font-size: 0.6rem;
    color: #a06060;
    margin-top: 0.3rem;
    letter-spacing: 2.5px;
    text-transform: uppercase;
}

.nav-section {
    font-family: 'DM Mono', monospace;
    font-size: 0.58rem;
    letter-spacing: 2px;
    text-transform: uppercase;
    color: #8a4a4a;
    padding: 1rem 1.2rem 0.3rem;
}

[data-testid="stSidebar"] .stButton > button {
    background: transparent !important;
    color: #DCC3AA !important;
    border: none !important;
    border-radius: 6px !important;
    font-family: 'DM Sans', sans-serif !important;
    font-weight: 500 !important;
    font-size: 0.88rem !important;
    padding: 0.55rem 1rem !important;
    text-align: left !important;
    width: 100% !important;
    transition: all 0.15s !important;
    box-shadow: none !important;
}
[data-testid="stSidebar"] .stButton > button:hover {
    background: #6b2a2a !important;
    color: #F1E2D1 !important;
    transform: none !important;
}

/* ── Page header ── */
.page-header {
    margin-bottom: 2.5rem;
    padding-bottom: 1.5rem;
    border-bottom: 2px solid #DCC3AA;
}
.page-eyebrow {
    font-family: 'DM Mono', monospace;
    font-size: 0.62rem;
    color: #810B38;
    letter-spacing: 3px;
    text-transform: uppercase;
    margin-bottom: 0.5rem;
}
.page-title {
    font-family: 'Playfair Display', serif;
    font-size: 2.2rem;
    font-weight: 900;
    color: #541A1A;
    letter-spacing: -1px;
    line-height: 1.1;
}
.page-title span { color: #810B38; }
.page-subtitle {
    font-family: 'DM Mono', monospace;
    font-size: 0.72rem;
    color: #a07060;
    margin-top: 0.4rem;
    letter-spacing: 1px;
}

/* ── Cards ── */
.kb-card {
    background: #FFFFFF;
    border: 1px solid #DCC3AA;
    border-radius: 10px;
    padding: 1.4rem 1.6rem;
    margin-bottom: 0.8rem;
    transition: border-color 0.2s, box-shadow 0.2s;
}
.kb-card:hover {
    border-color: #810B38;
    box-shadow: 0 2px 12px #810B3815;
}
.kb-card-title {
    font-family: 'DM Mono', monospace;
    font-size: 0.62rem;
    font-weight: 500;
    color: #810B38;
    letter-spacing: 2.5px;
    text-transform: uppercase;
    margin-bottom: 1rem;
}

/* ── Stats ── */
.stat-card {
    background: #FFFFFF;
    border: 1px solid #DCC3AA;
    border-radius: 10px;
    padding: 1.2rem 1.5rem;
}
.stat-value {
    font-family: 'Playfair Display', serif;
    font-size: 2rem;
    font-weight: 700;
    color: #541A1A;
}
.stat-label {
    font-family: 'DM Mono', monospace;
    font-size: 0.62rem;
    color: #a07060;
    letter-spacing: 2px;
    text-transform: uppercase;
    margin-top: 0.2rem;
}

/* ── Chunk items ── */
.chunk-item {
    background: #FFFFFF;
    border: 1px solid #DCC3AA;
    border-radius: 8px;
    padding: 1rem 1.3rem;
    margin-bottom: 0.5rem;
    transition: border-color 0.15s, box-shadow 0.15s;
}
.chunk-item:hover {
    border-color: #810B38;
    box-shadow: 0 2px 8px #810B3810;
}
.chunk-number {
    font-family: 'DM Mono', monospace;
    font-size: 0.62rem;
    color: #810B38;
    font-weight: 500;
    letter-spacing: 1px;
    margin-bottom: 0.4rem;
}
.chunk-text {
    font-size: 0.82rem;
    color: #541A1A;
    line-height: 1.65;
}
.chunk-meta {
    font-family: 'DM Mono', monospace;
    font-size: 0.6rem;
    color: #c0a090;
    margin-top: 0.5rem;
    letter-spacing: 1px;
}

/* ── Tags ── */
.kb-tag {
    display: inline-block;
    background: #810B3812;
    color: #810B38;
    border: 1px solid #810B3830;
    border-radius: 4px;
    padding: 2px 8px;
    font-size: 0.62rem;
    font-family: 'DM Mono', monospace;
    letter-spacing: 1px;
    text-transform: uppercase;
}
.kb-tag-green {
    background: #eafaf1;
    color: #27ae60;
    border: 1px solid #a9dfbf;
}

/* ── Info box ── */
.kb-info {
    background: #810B3808;
    border: 1px solid #810B3825;
    border-radius: 8px;
    padding: 0.75rem 1rem;
    font-size: 0.75rem;
    color: #810B38;
    margin-bottom: 1rem;
    font-family: 'DM Mono', monospace;
    letter-spacing: 0.5px;
}

/* ── Step indicator ── */
.step-indicator {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 2.5rem;
    padding: 1rem 1.5rem;
    background: #FFFFFF;
    border: 1px solid #DCC3AA;
    border-radius: 10px;
}
.step { display: flex; align-items: center; gap: 0.5rem; font-size: 0.72rem; font-family: 'DM Mono', monospace; }
.step-num {
    width: 24px; height: 24px; border-radius: 50%;
    background: #F1E2D1; color: #c0a090;
    display: flex; align-items: center; justify-content: center;
    font-size: 0.62rem; font-weight: 500; flex-shrink: 0;
}
.step-num.active { background: #810B38; color: #F1E2D1; }
.step-num.done   { background: #eafaf1; color: #27ae60; }
.step-label { color: #c0a090; white-space: nowrap; }
.step-label.active { color: #541A1A; }
.step-label.done   { color: #a0b0a0; }
.step-connector { flex: 1; height: 1px; background: #DCC3AA; min-width: 20px; max-width: 50px; }

/* ── Doc row ── */
.doc-row {
    background: #FFFFFF;
    border: 1px solid #DCC3AA;
    border-radius: 8px;
    padding: 1rem 1.4rem;
    margin-bottom: 0.5rem;
    display: flex;
    align-items: center;
    gap: 1rem;
    transition: border-color 0.15s, box-shadow 0.15s;
}
.doc-row:hover {
    border-color: #810B38;
    box-shadow: 0 2px 8px #810B3810;
}

/* ── Empty state ── */
.empty-state { text-align: center; padding: 5rem 2rem; }
.empty-icon  { font-size: 2.5rem; margin-bottom: 1rem; opacity: 0.25; }
.empty-title {
    font-family: 'Playfair Display', serif;
    font-size: 1.2rem; color: #c0a090; margin-bottom: 0.4rem;
}
.empty-sub {
    font-family: 'DM Mono', monospace;
    font-size: 0.65rem; color: #d0b8a0; letter-spacing: 1px;
}

/* ── Buttons ── */
.stButton > button {
    background: #810B38 !important;
    color: #F1E2D1 !important;
    border: none !important;
    border-radius: 8px !important;
    font-family: 'DM Sans', sans-serif !important;
    font-weight: 600 !important;
    font-size: 0.85rem !important;
    padding: 0.6rem 1.5rem !important;
    transition: all 0.2s !important;
}
.stButton > button:hover {
    background: #541A1A !important;
    transform: translateY(-1px) !important;
    box-shadow: 0 4px 12px #810B3830 !important;
}

/* ── Form elements ── */
.stSelectbox label, .stSlider label, .stRadio label {
    color: #a07060 !important;
    font-size: 0.72rem !important;
    font-family: 'DM Mono', monospace !important;
    letter-spacing: 1.5px !important;
    text-transform: uppercase !important;
}
.stSelectbox > div > div {
    background: #FFFFFF !important;
    border-color: #DCC3AA !important;
    color: #541A1A !important;
}
[data-testid="stFileUploader"] {
    background: #FFFFFF !important;
    border: 1px dashed #DCC3AA !important;
    border-radius: 10px !important;
}
.stTextInput > div > div > input {
    background: #FFFFFF !important;
    border-color: #DCC3AA !important;
    color: #541A1A !important;
    border-radius: 8px !important;
}

/* ── Scrollbar ── */
::-webkit-scrollbar { width: 4px; }
::-webkit-scrollbar-track { background: #F1E2D1; }
::-webkit-scrollbar-thumb { background: #DCC3AA; border-radius: 2px; }

[data-testid="column"] { padding: 0 0.5rem; }

/* ── Remove Streamlit default top padding ── */
.block-container {
    padding-top: 1.5rem !important;
}
[data-testid="stAppViewBlockContainer"] {
    padding-top: 1.5rem !important;
}
</style>
"""


def inject_theme():
    import streamlit as st

    st.markdown(THEME_CSS, unsafe_allow_html=True)


def sidebar_logo():
    import streamlit as st

    st.markdown(
        """
    <div class="kb-logo">
        <div class="kb-logo-text">Know<span>-Blocks</span></div>
        <div class="kb-logo-sub">RAG Pipeline Builder</div>
    </div>
    """,
        unsafe_allow_html=True,
    )


def page_header(eyebrow, title, accent, subtitle):
    import streamlit as st

    st.markdown(
        f"""
    <div class="page-header">
        <div class="page-eyebrow">{eyebrow}</div>
        <div class="page-title">{title} <span>{accent}</span></div>
        <div class="page-subtitle">{subtitle}</div>
    </div>
    """,
        unsafe_allow_html=True,
    )
