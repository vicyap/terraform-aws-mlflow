FROM python:3.8.2-slim

ARG MLFLOW_VERSION
ENV MLFLOW_VERSION=${MLFLOW_VERSION:-1.19.0}

RUN pip install PyMySQL==1.0.2 && \   
    pip install psycopg2-binary==2.9.1 && \
    pip install mlflow[extras]==$MLFLOW_VERSION

ENTRYPOINT ["mlflow", "server"]
