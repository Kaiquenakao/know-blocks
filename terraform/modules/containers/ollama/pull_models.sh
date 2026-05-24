#!/bin/bash
set -e

export OLLAMA_MODELS=/root/.ollama/models
mkdir -p $OLLAMA_MODELS

echo ">>> Starting Ollama temporarily..."
/bin/ollama serve &
OLLAMA_PID=$!

echo ">>> Waiting for Ollama to be ready..."
for i in $(seq 1 30); do
    if curl -sf http://localhost:11434 > /dev/null 2>&1; then
        echo ">>> Ollama ready."
        break
    fi
    echo ">>> Attempt $i/30..."
    sleep 2
done

echo ">>> Pulling nomic-embed-text..."
/bin/ollama pull nomic-embed-text

echo ">>> Pulling all-minilm..."
/bin/ollama pull all-minilm

echo ">>> Models saved:"
/bin/ollama list

echo ">>> Stopping temporary server..."
kill $OLLAMA_PID
wait $OLLAMA_PID 2>/dev/null || true
echo ">>> Done."
