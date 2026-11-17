# eProc TJSC — HTML de referência (consulta pública)

## Tentativa com CPF 03624841980 (Rafael Alberto Pasa) — 2026-11-13

| Resultado | Detalhe |
|---|---|
| Formulário público | **OK** — HTML real obtido via HTTP GET |
| POST com CPF | **Sem grid** — página volta com **Cloudflare Turnstile** |
| Browser (IP datacenter) | **WAF F5** — "Verificação de segurança do portal institucional" |
| Lista / capa / eventos | **Não capturados** nesta rodada |

### Bloqueios identificados

1. **Cloudflare Turnstile** no form (`cf-turnstile`, sitekey `0x4AAAAAAA9zgl8AcFbokEy4`). Sem token válido o POST não consulta.
2. **WAF F5** em automação browser a partir de IP de datacenter (OCI).

### Arquivos nesta pasta

| Arquivo | Conteúdo |
|---|---|
| `form_consulta_publica_fragment.html` | Fragmento real do `frmProcessoLista` + campos + Turnstile |
| `README.md` | Este mapa |

### URL canônica

```
https://eprocwebcon.tjsc.jus.br/consulta1g/externo_controlador.php?acao=processo_consulta_publica
```

### Campos úteis para o parser / driver

| name | id | uso |
|---|---|---|
| `txtNumProcesso` | `txtNumProcesso` | Nº CNJ |
| `txtNumChave` | `txtNumChave` | Chave processo |
| `txtNumChaveDocumento` | `txtNumChaveDocumento` | Chave documento |
| `txtStrParte` | `txtStrParte` | Nome da parte |
| `rdoTipo` | `rdoPessoaFisica` / `rdoPessoaJuridica` | `CPF` \| `CNPJ` |
| `txtCpfCnpj` | `txtCpfCnpj` | Documento |
| `txtStrOAB` | `txtStrOAB` | OAB |
| `sbmNovo` | `sbmNovo` | Submit |

### Como obter lista/capa de verdade (você no host)

```bash
# 1) Baixa formulário completo (HTTP)
chmod +x scripts/fetch_eproc_public.sh
./scripts/fetch_eproc_public.sh --cpf 03624841980

# 2) Browser com stealth (ainda precisa Turnstile/WAF ok)
go run ./cmd/capture_html -cpf 03624841980 -out testdata/eproc/public -headless=false
```

Recomendação operacional: capturar **uma vez** de rede residencial/escritório (ou sessão manual exportando OuterHTML da grid/capa) e versionar aqui em:

```text
lista_cpf_03624841980.html
capa_<numero_cnj>.html
eventos_<numero_cnj>.html
```

Parsers (só você) usam esses arquivos como fixtures.
