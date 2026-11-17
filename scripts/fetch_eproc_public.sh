#!/usr/bin/env bash
# Baixa o HTML da consulta pública eProc TJSC e grava em testdata/eproc/public/
# Uso:
#   ./scripts/fetch_eproc_public.sh
#   ./scripts/fetch_eproc_public.sh --cpf 03624841980
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="${ROOT}/testdata/eproc/public"
mkdir -p "${OUT_DIR}"

UA='Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36'
BASE='https://eprocwebcon.tjsc.jus.br/consulta1g'
FORM_URL="${BASE}/externo_controlador.php?acao=processo_consulta_publica"

CPF=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --cpf) CPF="$2"; shift 2 ;;
    *) echo "arg desconhecido: $1"; exit 1 ;;
  esac
done

COOKIE_JAR="$(mktemp)"
trap 'rm -f "${COOKIE_JAR}"' EXIT

echo ">> GET formulário"
curl -sS -L -A "${UA}" -H 'Accept-Language: pt-BR,pt;q=0.9' \
  -c "${COOKIE_JAR}" -b "${COOKIE_JAR}" \
  "${FORM_URL}" -o "${OUT_DIR}/consulta_publica_form.html"

echo "   salvo: ${OUT_DIR}/consulta_publica_form.html ($(wc -c < "${OUT_DIR}/consulta_publica_form.html") bytes)"

if [[ -n "${CPF}" ]]; then
  echo ">> POST consulta CPF=${CPF} (sem Turnstile — pode voltar só o form+captcha)"
  curl -sS -L -A "${UA}" \
    -c "${COOKIE_JAR}" -b "${COOKIE_JAR}" \
    -H 'Content-Type: application/x-www-form-urlencoded' \
    -H "Origin: https://eprocwebcon.tjsc.jus.br" \
    -H "Referer: ${FORM_URL}" \
    --data-urlencode "hdnInfraTipoPagina=1" \
    --data-urlencode "txtNumProcesso=" \
    --data-urlencode "txtNumChave=" \
    --data-urlencode "txtNumChaveDocumento=" \
    --data-urlencode "txtStrParte=" \
    --data-urlencode "chkFonetica=N" \
    --data-urlencode "rdoTipo=CPF" \
    --data-urlencode "txtCpfCnpj=${CPF}" \
    --data-urlencode "txtStrOAB=" \
    --data-urlencode "hdnInfraCaptcha=0" \
    --data-urlencode "hdnInfraSelecoes=Infra" \
    --data-urlencode "hdnInfraPrefixoCookie=TJSC_Eproc_" \
    --data-urlencode "sbmNovo=Consultar" \
    "${BASE}/externo_controlador.php?acao=processo_consulta_publica&acao_origem=principal&acao_retorno=processo_consulta_publica" \
    -o "${OUT_DIR}/consulta_publica_post_cpf_${CPF}.html"
  echo "   salvo: ${OUT_DIR}/consulta_publica_post_cpf_${CPF}.html"
fi

echo "OK. Para lista/capa reais use o browser com Turnstile resolvido: go run ./cmd/capture_html -cpf ${CPF:-SEU_CPF}"
