import streamlit as st
import sys, os, datetime

sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))
from theme import inject_theme, sidebar_logo, page_header

st.set_page_config(
    page_title="Deploy · Know-Blocks",
    page_icon="⬆",
    layout="wide",
    initial_sidebar_state="expanded",
)

inject_theme()

# ── Session state defaults ────────────────────────────────────────────────────
for key, val in [
    ("documents", []),
    ("total_chunks", 0),
    ("collections", []),
    ("deploy_step", 1),
    ("_chunk_strategy", "Fixed Size"),
    ("_embed_model", "text-embedding-3-small (OpenAI)"),
    ("_chunk_size", 512),
    ("_overlap", 64),
    ("_pending_files", []),
]:
    if key not in st.session_state:
        st.session_state[key] = val

# ── Sidebar ───────────────────────────────────────────────────────────────────
with st.sidebar:
    sidebar_logo()
    st.markdown('<div class="nav-section">Workspace</div>', unsafe_allow_html=True)
    st.page_link("app.py", label="🏠  Document Store")
    st.page_link("pages/deploy.py", label="⬆  Deploy")
    st.page_link("pages/search.py", label="🔍  Search Chunks")
    st.markdown('<div class="nav-section">Manage</div>', unsafe_allow_html=True)
    st.page_link("pages/collections.py", label="🗂  Collections")
    st.page_link("pages/settings.py", label="⚙  Settings")

    st.markdown("<br>" * 4, unsafe_allow_html=True)
    docs = len(st.session_state.documents)
    chunks = st.session_state.total_chunks
    st.markdown(
        f"""
    <div style="border-top:1px solid #3d1515; padding:1rem 0.5rem;">
        <div style="font-family:'DM Mono',monospace; font-size:0.58rem; color:#4a1a1a;
                    letter-spacing:2px; text-transform:uppercase; margin-bottom:0.8rem;">Platform</div>
        <div style="display:flex; justify-content:space-between; margin-bottom:0.35rem;">
            <span style="font-size:0.78rem; color:#6b3030;">Documents</span>
            <span style="font-family:'DM Mono',monospace; font-size:0.78rem; color:#DCC3AA;">{docs}</span>
        </div>
        <div style="display:flex; justify-content:space-between;">
            <span style="font-size:0.78rem; color:#6b3030;">Chunks</span>
            <span style="font-family:'DM Mono',monospace; font-size:0.78rem; color:#DCC3AA;">{chunks}</span>
        </div>
    </div>
    """,
        unsafe_allow_html=True,
    )

# ── Main ──────────────────────────────────────────────────────────────────────
page_header(
    "Know-Blocks · Deploy",
    "Document",
    "Deploy",
    "// upload → configure chunking → preview → deploy to store",
)

step = st.session_state.deploy_step

# Step indicator
steps_meta = [(1, "Upload"), (2, "Configure"), (3, "Preview"), (4, "Done")]
html = '<div class="step-indicator">'
for i, (num, label) in enumerate(steps_meta):
    cls = "done" if num < step else ("active" if num == step else "")
    icon = "✓" if num < step else str(num)
    lcls = "done" if num < step else ("active" if num == step else "")
    html += f'<div class="step"><div class="step-num {cls}">{icon}</div><span class="step-label {lcls}">{label}</span></div>'
    if i < len(steps_meta) - 1:
        html += '<div class="step-connector"></div>'
html += "</div>"
st.markdown(html, unsafe_allow_html=True)

# ══════════════════════════════════════════════════════════════════════════════
# STEP 1 — Upload
# ══════════════════════════════════════════════════════════════════════════════
if step == 1:
    st.markdown(
        '<div class="kb-card-title">Upload Documents</div>', unsafe_allow_html=True
    )

    uploaded = st.file_uploader(
        "Drop files here",
        type=["pdf", "txt", "docx", "md"],
        accept_multiple_files=True,
        label_visibility="collapsed",
    )

    st.markdown(
        """
    <div class="kb-info">
        ▸ Supported formats: PDF · TXT · DOCX · Markdown &nbsp;·&nbsp; Max 50 MB per file
    </div>
    """,
        unsafe_allow_html=True,
    )

    if uploaded:
        st.markdown(
            '<div class="kb-card-title" style="margin-top:1.5rem;">Ready to configure</div>',
            unsafe_allow_html=True,
        )
        for f in uploaded:
            size_kb = round(f.size / 1024, 1)
            ext = f.name.split(".")[-1].upper()
            st.markdown(
                f"""
            <div class="doc-row">
                <div style="font-size:1.5rem; opacity:0.6;">📄</div>
                <div style="flex:1;">
                    <div style="font-weight:600; color:#F1E2D1; font-size:0.88rem;">{f.name}</div>
                    <div style="font-family:'DM Mono',monospace; font-size:0.62rem;
                                color:#6b3030; margin-top:0.2rem;">{size_kb} KB</div>
                </div>
                <span class="kb-tag">{ext}</span>
            </div>
            """,
                unsafe_allow_html=True,
            )

        st.session_state._pending_files = uploaded
        st.markdown("<div style='margin-top:1.2rem'></div>", unsafe_allow_html=True)
        if st.button("Configure chunking →"):
            st.session_state.deploy_step = 2
            st.rerun()

# ══════════════════════════════════════════════════════════════════════════════
# STEP 2 — Configure
# ══════════════════════════════════════════════════════════════════════════════
elif step == 2:
    col_left, col_right = st.columns([1.2, 1], gap="large")

    with col_left:
        st.markdown(
            '<div class="kb-card-title">Chunking Strategy</div>', unsafe_allow_html=True
        )
        strategy = st.selectbox(
            "Method",
            ["Fixed Size", "Sentence-based", "Paragraph-based", "Semantic"],
        )

        if strategy == "Fixed Size":
            chunk_size = st.slider("Chunk size (tokens)", 128, 1024, 512, step=64)
            overlap = st.slider("Overlap (tokens)", 0, 256, 64, step=16)
            st.markdown(
                f"""
            <div class="kb-info">▸ Each chunk ~{chunk_size} tokens · Overlap {overlap} tokens</div>
            """,
                unsafe_allow_html=True,
            )
            st.session_state._chunk_size = chunk_size
            st.session_state._overlap = overlap

        elif strategy == "Sentence-based":
            n_sent = st.slider("Sentences per chunk", 2, 10, 4)
            st.markdown(
                f'<div class="kb-info">▸ Groups {n_sent} sentences per chunk</div>',
                unsafe_allow_html=True,
            )

        elif strategy == "Paragraph-based":
            st.markdown(
                '<div class="kb-info">▸ Splits on paragraph breaks — great for structured docs</div>',
                unsafe_allow_html=True,
            )

        elif strategy == "Semantic":
            thresh = st.slider("Similarity threshold", 0.50, 0.99, 0.85, step=0.01)
            st.markdown(
                f'<div class="kb-info">▸ Groups semantically similar sentences · threshold {thresh}</div>',
                unsafe_allow_html=True,
            )

        st.session_state._chunk_strategy = strategy

    with col_right:
        st.markdown(
            '<div class="kb-card-title">Embedding Model</div>', unsafe_allow_html=True
        )
        st.markdown(
            """
        <div class="kb-info">▸ Open source · runs locally via Ollama · no API cost</div>
        """,
            unsafe_allow_html=True,
        )

        model = st.selectbox(
            "Default model",
            ["nomic-embed-text", "all-minilm"],
            format_func=lambda m: {
                "nomic-embed-text": "nomic-embed-text — 768 dims · 274 MB",
                "all-minilm": "all-minilm — 384 dims · 45 MB",
            }[m],
        )
        st.session_state._embed_model = model

        info = {
            "nomic-embed-text": (
                "768",
                "free · Ollama",
                "recommended",
                "Best quality/size ratio — ideal for most use cases",
            ),
            "all-minilm": (
                "384",
                "free · Ollama",
                "lightweight",
                "Ultra fast — great for large document sets",
            ),
        }
        dims, cost, note, desc = info[model]
        st.markdown(
            f"""
        <div class="kb-card" style="margin-top:0.5rem;">
            <div style="display:flex; justify-content:space-between; margin-bottom:0.7rem;">
                <span style="font-size:0.78rem; color:#a07060;">Dimensions</span>
                <span style="font-family:'DM Mono',monospace; font-size:0.78rem; color:#541A1A;">{dims}</span>
            </div>
            <div style="display:flex; justify-content:space-between; margin-bottom:0.7rem;">
                <span style="font-size:0.78rem; color:#a07060;">Cost</span>
                <span style="font-family:'DM Mono',monospace; font-size:0.78rem; color:#27ae60;">{cost}</span>
            </div>
            <div style="display:flex; justify-content:space-between; margin-bottom:0.9rem;">
                <span style="font-size:0.78rem; color:#a07060;">Profile</span>
                <span class="kb-tag">{note}</span>
            </div>
            <div style="font-size:0.75rem; color:#a07060; line-height:1.5; border-top:1px solid #DCC3AA;
                        padding-top:0.7rem; font-family:'DM Mono',monospace;">
                {desc}
            </div>
        </div>
        """,
            unsafe_allow_html=True,
        )

    st.markdown("<div style='margin-top:1.5rem'></div>", unsafe_allow_html=True)
    col_back, col_next, _ = st.columns([1, 2, 4])
    with col_back:
        if st.button("← Back"):
            st.session_state.deploy_step = 1
            st.rerun()
    with col_next:
        if st.button("Preview chunks →"):
            st.session_state.deploy_step = 3
            st.rerun()

# ══════════════════════════════════════════════════════════════════════════════
# STEP 3 — Preview + Metadata
# ══════════════════════════════════════════════════════════════════════════════
elif step == 3:
    strategy = st.session_state._chunk_strategy
    model = st.session_state._embed_model

    col_meta, col_preview = st.columns([1, 1.4], gap="large")

    # ── Left: Metadata form ───────────────────────────────────────────────────
    with col_meta:
        st.markdown(
            '<div class="kb-card-title">Document Metadata</div>', unsafe_allow_html=True
        )
        st.markdown(
            """
        <div class="kb-info">
            ▸ These fields are stored alongside the document and searchable in the store
        </div>
        """,
            unsafe_allow_html=True,
        )

        doc_title = st.text_input(
            "Title",
            placeholder="e.g. Q4 2024 Annual Report",
            help="Human-readable name for this document",
        )
        description = st.text_area(
            "Description",
            placeholder="Brief description of this document's content and purpose...",
            height=90,
            help="Helps curators understand what this document covers",
        )

        col_a, col_b = st.columns(2)
        with col_a:
            language = st.selectbox(
                "Language",
                ["Portuguese", "English", "Spanish", "French", "German", "Other"],
            )
        with col_b:
            doc_type = st.selectbox(
                "Document type",
                ["Report", "Manual", "FAQ", "Policy", "Contract", "Research", "Other"],
            )

        tags_raw = st.text_input(
            "Tags",
            placeholder="e.g. finance, 2024, quarterly",
            help="Comma-separated tags for filtering",
        )
        tags = [t.strip() for t in tags_raw.split(",") if t.strip()] if tags_raw else []

        source = st.text_input(
            "Source / Author",
            placeholder="e.g. McKinsey Global Institute",
        )

        visibility = st.radio(
            "Visibility",
            ["Private", "Public"],
            horizontal=True,
            help="Public documents are visible to all users in the platform",
        )

        # Summary card
        st.markdown("<div style='margin-top:1rem'></div>", unsafe_allow_html=True)
        tags_html = (
            "".join(
                [
                    f'<span class="kb-tag" style="margin-bottom:4px;">{t}</span>'
                    for t in tags
                ]
            )
            or "—"
        )
        st.markdown(
            f"""
        <div class="kb-card">
            <div class="kb-card-title">Summary</div>
            <div style="display:flex; justify-content:space-between; margin-bottom:0.5rem;">
                <span style="font-size:0.78rem; color:#a07060;">Model</span>
                <span style="font-family:'DM Mono',monospace; font-size:0.75rem; color:#810B38;">{model}</span>
            </div>
            <div style="display:flex; justify-content:space-between; margin-bottom:0.5rem;">
                <span style="font-size:0.78rem; color:#a07060;">Strategy</span>
                <span style="font-family:'DM Mono',monospace; font-size:0.75rem; color:#541A1A;">{strategy}</span>
            </div>
            <div style="display:flex; justify-content:space-between; margin-bottom:0.5rem;">
                <span style="font-size:0.78rem; color:#a07060;">Language</span>
                <span style="font-family:'DM Mono',monospace; font-size:0.75rem; color:#541A1A;">{language}</span>
            </div>
            <div style="display:flex; justify-content:space-between; margin-bottom:0.5rem;">
                <span style="font-size:0.78rem; color:#a07060;">Type</span>
                <span style="font-family:'DM Mono',monospace; font-size:0.75rem; color:#541A1A;">{doc_type}</span>
            </div>
            <div style="display:flex; justify-content:space-between; margin-bottom:0.5rem;">
                <span style="font-size:0.78rem; color:#a07060;">Visibility</span>
                <span class="kb-tag">{visibility}</span>
            </div>
            <div style="margin-top:0.7rem; border-top:1px solid #DCC3AA; padding-top:0.7rem;">
                <span style="font-size:0.72rem; color:#a07060;">Tags &nbsp;</span>
                {tags_html}
            </div>
        </div>
        """,
            unsafe_allow_html=True,
        )

    # ── Right: Chunk preview ──────────────────────────────────────────────────
    with col_preview:
        st.markdown(
            '<div class="kb-card-title">Chunk Sampling</div>', unsafe_allow_html=True
        )
        st.markdown(
            """
        <div class="kb-info">▸ Real chunks extracted from your document</div>
        """,
            unsafe_allow_html=True,
        )

        # ── Real chunking ─────────────────────────────────────────────────────
        sample_chunks = []
        files = st.session_state.get("_pending_files", [])

        if files:
            try:
                import fitz  # pymupdf
                import textwrap

                f = files[0]
                raw = f.read()
                f.seek(0)  # reset pointer

                doc_pdf = fitz.open(stream=raw, filetype="pdf")
                full_text = ""
                for page in doc_pdf:
                    full_text += page.get_text()
                doc_pdf.close()

                # Chunking real por tamanho fixo
                words = full_text.split()
                size = st.session_state.get("_chunk_size", 512)
                ovlp = st.session_state.get("_overlap", 64)

                i = 0
                while i < len(words) and len(sample_chunks) < 6:
                    chunk_words = words[i : i + size]
                    chunk_text = " ".join(chunk_words).strip()
                    if len(chunk_text) > 30:
                        sample_chunks.append(chunk_text)
                    i += max(1, size - ovlp)

            except ImportError:
                st.markdown(
                    """
                <div class="kb-info" style="color:#c0392b; border-color:#c0392b33; background:#fff5f5;">
                    ▸ Install pymupdf to enable real chunking: <code>pip install pymupdf</code>
                </div>
                """,
                    unsafe_allow_html=True,
                )
            except Exception as e:
                st.markdown(
                    f"""
                <div class="kb-info" style="color:#c0392b;">▸ Could not parse file: {e}</div>
                """,
                    unsafe_allow_html=True,
                )

        # Fallback se não conseguiu extrair
        if not sample_chunks:
            sample_chunks = [
                "No text could be extracted. Make sure the file is a text-based PDF (not scanned).",
            ]

        for i, text in enumerate(sample_chunks):
            tokens_est = int(len(text.split()) * 1.3)
            is_truncated = len(text) > 280
            preview = text[:280] + ("..." if is_truncated else "")

            st.markdown(
                f"""
            <div class="chunk-item">
                <div class="chunk-number">CHUNK_{str(i).zfill(3)}</div>
                <div class="chunk-text">{preview}</div>
                <div class="chunk-meta">~{tokens_est} tokens · {strategy} · {model}</div>
            </div>
            """,
                unsafe_allow_html=True,
            )

            if is_truncated:
                with st.expander("ver completo"):
                    st.markdown(
                        f"""
                    <div style="font-size:0.82rem; color:#541A1A; line-height:1.75;
                                font-family:'DM Sans',sans-serif; white-space:pre-wrap;">{text}</div>
                    """,
                        unsafe_allow_html=True,
                    )
                    st.caption(f"~{tokens_est} tokens · {len(text)} chars")

        st.markdown(
            f"""
        <div class="kb-card" style="margin-top:1rem; display:flex; gap:2.5rem; flex-wrap:wrap;">
            <div>
                <div class="stat-label">Sampled chunks</div>
                <div style="font-family:'Playfair Display',serif; font-size:1.5rem;
                            font-weight:700; color:#541A1A; margin-top:0.2rem;">{len(sample_chunks)}</div>
            </div>
            <div>
                <div class="stat-label">Storage</div>
                <div style="font-size:0.8rem; font-weight:600; color:#541A1A;
                            margin-top:0.2rem; font-family:'DM Mono',monospace;">S3 + S3 Vectors</div>
            </div>
        </div>
        """,
            unsafe_allow_html=True,
        )

    # ── Navigation ────────────────────────────────────────────────────────────
    st.markdown("<div style='margin-top:2rem'></div>", unsafe_allow_html=True)
    col_back, col_next, _ = st.columns([1, 2, 4])
    with col_back:
        if st.button("← Back"):
            st.session_state.deploy_step = 2
            st.rerun()
    with col_next:
        if st.button("🚀 Deploy to Store"):
            files = st.session_state._pending_files or []
            import zoneinfo

            now = datetime.datetime.now(
                tz=zoneinfo.ZoneInfo("America/Sao_Paulo")
            ).strftime("%d/%m/%Y %H:%M:%S")
            for f in files if files else [{"name": "document.pdf", "size": 0}]:
                name = f.name if hasattr(f, "name") else f["name"]
                st.session_state.documents.append(
                    {
                        "name": doc_title or name,
                        "filename": name,
                        "description": description,
                        "language": language,
                        "doc_type": doc_type,
                        "tags": tags,
                        "source": source,
                        "visibility": visibility,
                        "chunks": len(sample_chunks),
                        "model": model,
                        "strategy": strategy,
                        "embeddings": 0,
                        "deployed_at": now,
                    }
                )
            st.session_state.total_chunks += len(sample_chunks) * max(
                1, len(files) if files else 1
            )
            st.session_state.deploy_step = 4
            st.rerun()

# ══════════════════════════════════════════════════════════════════════════════
# STEP 4 — Done
# ══════════════════════════════════════════════════════════════════════════════
elif step == 4:
    doc_count = len(st.session_state.documents)
    st.markdown(
        f"""
    <div style="text-align:center; padding:3.5rem 2rem;">
        <div style="font-size:2.8rem; margin-bottom:1rem;">✓</div>
        <div style="font-family:'Playfair Display',serif; font-size:2rem; font-weight:900;
                    color:#F1E2D1; margin-bottom:0.5rem;">Deployed successfully</div>
        <div style="font-family:'DM Mono',monospace; font-size:0.72rem; color:#6b3030;
                    letter-spacing:1px; margin-bottom:0.3rem;">
            Document is now in the store · ready to be searched and added to collections
        </div>
        <div style="font-family:'DM Mono',monospace; font-size:0.65rem; color:#4a1a1a;
                    letter-spacing:1px; margin-bottom:2.5rem;">
            {doc_count} document{"s" if doc_count != 1 else ""} in store
        </div>
    </div>
    """,
        unsafe_allow_html=True,
    )

    col1, col2, col3, _ = st.columns([1.2, 1.2, 1.2, 2])
    with col1:
        if st.button("⬆  Deploy another", use_container_width=True):
            st.session_state.deploy_step = 1
            st.rerun()
    with col2:
        if st.button("🏠  View Store", use_container_width=True):
            st.session_state.deploy_step = 1
            st.switch_page("app.py")
    with col3:
        if st.button("🔍  Search Chunks", use_container_width=True):
            st.session_state.deploy_step = 1
            st.switch_page("pages/search.py")
