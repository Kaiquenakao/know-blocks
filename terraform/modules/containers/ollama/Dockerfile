FROM ollama/ollama:latest

# Copia o script de pull
COPY pull_models.sh /pull_models.sh
RUN chmod +x /pull_models.sh

# Define onde os modelos ficam salvos
ENV OLLAMA_MODELS=/root/.ollama/models

# Puxa os modelos durante o build e salva no path correto
RUN /pull_models.sh

EXPOSE 11434
ENTRYPOINT ["/bin/ollama"]
CMD ["serve"]
