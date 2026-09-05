# Task 4: Validação Integrada, Auditoria de Qualidade e Deploy Test

## Contexto & Localização
Repositório: `/root/chaossql`
Diretório do site: `/root/chaossql/site`

## Objetivos Exatos
1. Validação Sintática Completa:
   - `node -c site/docs-data.js`
   - `node -c site/app.js`
2. Validação da Suíte Go do Repositório:
   - `go test ./...`
3. Teste de Servidor HTTP Local:
   - Iniciar servidor Python em porta efêmera (ex: 8089)
   - Executar `curl -I` para `/`, `/docs-data.js`, `/app.js`, `/assets/style.css`
4. Auditoria de Coerência Linguística nos 2 Modos:
   - Verificar que no modo `pt`: todos os textos da UI, cards de engenharia, descrições dos cenários e capítulos estejam em português.
   - Verificar que no modo `en`: nenhum texto de UI ou parágrafo explicativo permaneça em português.
   - Testar o comportamento de alternância rápida no cliente.
5. Elaboração do Relatório Final de QA em `task-4-report.md`.
