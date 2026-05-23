"""
Know-Blocks — API Client
Utilitários para chamar o API Gateway a partir do Streamlit.
"""

import json
import datetime
import zoneinfo
import httpx
from utils.config import ENDPOINT_DEPLOY, ENDPOINT_STATUS, TIMEOUT_UPLOAD, TIMEOUT_STATUS


# ── Tipos de resposta ─────────────────────────────────────────────────────────

class PrepareResponse:
    def __init__(self, data: dict):
        self.success      = data.get("statusCode", 200) not in (409, 400, 500)
        self.upload_id    = data.get("upload_id", "")
        self.presigned_url = data.get("presigned_url", "")
        self.s3_key       = data.get("s3_key", "")
        self.error        = data.get("error", "")
        self.message      = data.get("message", "")
        self.status_code  = data.get("statusCode", 200)

    @property
    def is_conflict(self):
        return self.status_code == 409


class ConfirmResponse:
    def __init__(self, data: dict):
        self.success       = data.get("statusCode", 200) == 200
        self.execution_arn = data.get("execution_arn", "")
        self.s3_key        = data.get("s3_key", "")
        self.deployed_at   = data.get("deployed_at", "")
        self.error         = data.get("error", "")
        self.message       = data.get("message", "")
        self.status_code   = data.get("statusCode", 200)


class StatusResponse:
    def __init__(self, data: dict):
        self.step       = data.get("step", "")
        self.status     = data.get("status", "")
        self.progress   = data.get("progress", 0)
        self.message    = data.get("message", "")
        self.error      = data.get("error", "")
        self.updated_at = data.get("updated_at", "")

    @property
    def is_done(self):
        return self.status == "SUCCEEDED"

    @property
    def is_error(self):
        return self.status in ("FAILED", "error")

    @property
    def is_running(self):
        return self.status in ("RUNNING", "running")


# ── Step 1: Prepare — valida + gera presigned URL ────────────────────────────

def prepare_deploy(filename: str) -> PrepareResponse:
    """
    GET /documents/deploy?filename=X
    Valida se o arquivo já existe e retorna presigned URL para upload direto no S3.
    """
    try:
        r = httpx.get(
            ENDPOINT_DEPLOY,
            params={"filename": filename},
            timeout=10,
        )

        print(f"[prepare] status={r.status_code} body={r.text[:200]}")

        data = _parse_response(r)
        return PrepareResponse(data)

    except httpx.TimeoutException:
        return PrepareResponse({"statusCode": 504, "error": "timeout", "message": "Timeout ao preparar deploy."})
    except httpx.HTTPError as e:
        return PrepareResponse({"statusCode": 500, "error": "http_error", "message": str(e)})


# ── Step 2: Upload — envia PDF direto pro S3 via presigned URL ────────────────

def upload_to_s3(presigned_url: str, file) -> bool:
    """
    PUT presigned_url
    Envia o PDF direto pro S3 — bypassa o API Gateway.
    Retorna True se sucesso.
    """
    try:
        file.seek(0)
        content = file.read()

        r = httpx.put(
            presigned_url,
            content=content,
            headers={"Content-Type": "application/pdf"},
            timeout=TIMEOUT_UPLOAD,
        )

        print(f"[upload_s3] status={r.status_code}")
        return r.status_code in (200, 204)

    except httpx.HTTPError as e:
        print(f"[upload_s3] erro: {e}")
        return False


# ── Step 3: Confirm — inicia Step Function ────────────────────────────────────

def confirm_deploy(
    upload_id: str,
    s3_key: str,
    filename: str,
    chunk_strategy: str,
    chunk_size: int,
    overlap: int,
    embed_model: str,
    title: str,
    description: str,
    language: str,
    doc_type: str,
    tags: list,
    source: str,
    visibility: str,
) -> ConfirmResponse:
    """
    POST /documents/deploy
    Confirma o upload e inicia a Step Function.
    """
    now = datetime.datetime.now(
        tz=zoneinfo.ZoneInfo("America/Sao_Paulo")
    ).strftime("%d/%m/%Y %H:%M:%S")

    payload = {
        "document": {
            "filename":  filename,
            "s3_key":    s3_key,
            "mime_type": "application/pdf",
        },
        "chunking": {
            "method": _normalize_strategy(chunk_strategy),
            "params": {
                "chunk_size": chunk_size,
                "overlap":    overlap,
            },
        },
        "embedding": {
            "model":      embed_model,
            "dimensions": 768 if embed_model == "nomic-embed-text" else 384,
            "provider":   "ollama",
        },
        "metadata": {
            "title":       title or filename,
            "description": description,
            "language":    language,
            "doc_type":    doc_type,
            "tags":        tags,
            "source":      source,
            "visibility":  visibility.lower(),
        },
        "deployed_at": now,
    }

    body = {
        "upload_id": upload_id,
        "s3_key":    s3_key,
        "payload":   payload,
    }
    print(f"[confirm] enviando body={json.dumps(body)[:300]}")

    try:
        r = httpx.post(
            ENDPOINT_DEPLOY,
            json=body,
            timeout=30,
        )

        print(f"[confirm] status={r.status_code} body={r.text[:200]}")

        data = _parse_response(r)
        return ConfirmResponse(data)

    except httpx.TimeoutException:
        return ConfirmResponse({"statusCode": 504, "error": "timeout", "message": "Timeout ao confirmar deploy."})
    except httpx.HTTPError as e:
        return ConfirmResponse({"statusCode": 500, "error": "http_error", "message": str(e)})


# ── Status / Polling ──────────────────────────────────────────────────────────

def get_execution_status(execution_arn: str) -> StatusResponse:
    import urllib.parse
    encoded = urllib.parse.quote(execution_arn, safe="")

    try:
        r = httpx.get(
            f"{ENDPOINT_STATUS}/{encoded}",
            timeout=TIMEOUT_STATUS,
        )
        data = _parse_response(r)
        return StatusResponse(data)

    except httpx.TimeoutException:
        return StatusResponse({"status": "error", "error": "timeout"})
    except httpx.HTTPError as e:
        return StatusResponse({"status": "error", "error": str(e)})


# ── Helpers ───────────────────────────────────────────────────────────────────

def _parse_response(r: httpx.Response) -> dict:
    if not r.content:
        return {"statusCode": r.status_code, "error": "empty_response", "message": f"HTTP {r.status_code}"}
    try:
        data = r.json()
        if isinstance(data.get("body"), str):
            data = json.loads(data["body"])
        return data
    except Exception:
        return {"statusCode": r.status_code, "error": "parse_error", "message": r.text[:200]}


def _normalize_strategy(strategy: str) -> str:
    mapping = {
        "Fixed Size":      "fixed_size",
        "Sentence-based":  "sentence",
        "Paragraph-based": "paragraph",
        "Semantic":        "semantic",
    }
    return mapping.get(strategy, strategy.lower().replace(" ", "_"))