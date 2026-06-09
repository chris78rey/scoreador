package webapp

const pageTemplateWeb = `
<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Scoreador - Partido unico y lote Excel</title>
  <style>
    :root {
      --bg: #0f172a;
      --panel: #111827;
      --panel-2: #1f2937;
      --text: #e5e7eb;
      --muted: #9ca3af;
      --accent: #22c55e;
      --accent-2: #38bdf8;
      --accent-3: #f59e0b;
      --accent-4: #a78bfa;
      --line: #334155;
      --danger: #f87171;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: Segoe UI, Arial, sans-serif;
      background: radial-gradient(circle at top, #1d4ed8 0%, var(--bg) 50%);
      color: var(--text);
    }
    .wrap {
      max-width: 1240px;
      margin: 0 auto;
      padding: 24px;
    }
    .hero {
      display: grid;
      gap: 12px;
      grid-template-columns: 1.3fr 0.7fr;
      align-items: end;
      margin-bottom: 24px;
    }
    .card {
      background: rgba(17, 24, 39, 0.92);
      border: 1px solid rgba(148, 163, 184, 0.18);
      border-radius: 18px;
      padding: 20px;
      box-shadow: 0 20px 40px rgba(0,0,0,0.25);
      backdrop-filter: blur(10px);
    }
    h1, h2, h3, h4 { margin: 0 0 12px; }
    h1 { font-size: 30px; }
    p { color: var(--muted); margin: 0; line-height: 1.5; }
    .grid {
      display: grid;
      gap: 18px;
      grid-template-columns: 1fr 1fr;
    }
    .form-grid {
      display: grid;
      gap: 12px;
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
    .field { display: grid; gap: 6px; }
    .field-help {
      color: var(--muted);
      font-size: 12px;
      line-height: 1.35;
    }
    label { font-size: 12px; color: var(--muted); text-transform: uppercase; letter-spacing: .06em; }
    input, select {
      width: 100%;
      padding: 11px 12px;
      border-radius: 12px;
      border: 1px solid var(--line);
      background: var(--panel-2);
      color: var(--text);
      outline: none;
    }
    input:focus, select:focus { border-color: var(--accent-2); }
    .full { grid-column: 1 / -1; }
    .actions { display: flex; gap: 12px; align-items: center; margin-top: 14px; }
    button {
      border: 0;
      background: linear-gradient(135deg, var(--accent), #16a34a);
      color: #052e16;
      font-weight: 700;
      padding: 12px 16px;
      border-radius: 12px;
      cursor: pointer;
    }
    .hint { color: var(--muted); font-size: 13px; }
    .error {
      margin-bottom: 18px;
      padding: 12px 14px;
      border-radius: 12px;
      background: rgba(248, 113, 113, .12);
      border: 1px solid rgba(248, 113, 113, .35);
      color: #fecaca;
    }
    .result-grid {
      display: grid;
      gap: 12px;
      grid-template-columns: repeat(4, minmax(0, 1fr));
    }
    .metric {
      background: rgba(31, 41, 55, 0.8);
      border: 1px solid rgba(148, 163, 184, 0.14);
      border-radius: 14px;
      padding: 14px;
    }
    .metric span { display: block; color: var(--muted); font-size: 12px; text-transform: uppercase; letter-spacing: .06em; }
    .metric strong { display: block; margin-top: 6px; font-size: 20px; }
    .metric small { display: block; margin-top: 4px; color: var(--muted); font-size: 12px; }
    .split-grid {
      display: grid;
      gap: 16px;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      margin-top: 16px;
    }
    .score-grid {
      display: grid;
      gap: 12px;
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
    .score-card {
      background: rgba(15, 23, 42, 0.55);
      border: 1px solid rgba(148, 163, 184, 0.14);
      border-radius: 14px;
      padding: 14px;
    }
    .score-card h4 {
      margin: 0 0 8px;
      font-size: 18px;
    }
    .score-card .score-meta {
      display: flex;
      justify-content: space-between;
      gap: 10px;
      color: var(--muted);
      font-size: 12px;
      margin-bottom: 8px;
    }
    .score-bar {
      height: 10px;
      border-radius: 999px;
      overflow: hidden;
      background: rgba(148, 163, 184, 0.12);
    }
    .score-bar > span {
      display: block;
      height: 100%;
      border-radius: inherit;
      background: linear-gradient(90deg, var(--accent), var(--accent-2));
    }
    .summary-pill {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      padding: 8px 12px;
      border-radius: 999px;
      background: rgba(56, 189, 248, 0.12);
      color: #bae6fd;
      margin-bottom: 12px;
      font-size: 13px;
    }
    table {
      width: 100%;
      border-collapse: collapse;
      overflow: hidden;
      border-radius: 14px;
    }
    th, td {
      padding: 10px 12px;
      border-bottom: 1px solid rgba(148, 163, 184, 0.14);
      text-align: left;
      vertical-align: top;
    }
    th { color: var(--muted); font-size: 12px; text-transform: uppercase; letter-spacing: .06em; }
    tr:hover td { background: rgba(59, 130, 246, 0.06); }
    .badge {
      display: inline-block;
      padding: 4px 8px;
      border-radius: 999px;
      background: rgba(56, 189, 248, .12);
      color: #bae6fd;
      font-size: 12px;
    }
    .download-link {
      display: inline-flex;
      align-items: center;
      padding: 10px 14px;
      border-radius: 12px;
      background: rgba(34, 197, 94, 0.16);
      color: #bbf7d0;
      text-decoration: none;
      border: 1px solid rgba(34, 197, 94, 0.3);
    }
    .download-link:hover { background: rgba(34, 197, 94, 0.24); }
    .section { margin-top: 18px; }
    .footer-note { color: var(--muted); font-size: 12px; margin-top: 10px; }
    @media (max-width: 900px) {
      .hero, .grid, .result-grid, .form-grid, .split-grid, .score-grid { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <div class="wrap">
    <div class="hero">
      <div class="card">
        <h1>Simulador de partido unico</h1>
        <p>Corre una serie de simulaciones, guarda el resultado en SQLite y revisa el historial desde el navegador.</p>
      </div>
        <div class="card">
        <span class="badge">GUI local</span>
        <p class="footer-note">Abre esta pagina en tu navegador y ajusta equipos, tiros, motivacion 0-10, semilla y cantidad de corridas.</p>
      </div>
    </div>

    {{if .Error}}
    <div class="error">{{.Error}}</div>
    {{end}}

    {{if .BatchError}}
    <div class="error">{{.BatchError}}</div>
    {{end}}
    {{if .BatchMessage}}
    <div class="card section">
      <span class="badge">Excel procesado</span>
      <p>{{.BatchMessage}}</p>
      {{if .BatchDownloadURL}}
      <div class="actions" style="margin-top: 12px;">
        <a class="download-link" href="{{.BatchDownloadURL}}">{{if .BatchDownloadName}}Descargar {{.BatchDownloadName}}{{else}}Descargar resultado{{end}}</a>
      </div>
      {{end}}
    </div>
    {{end}}

    <div class="card section">
      <h2>Procesar Excel con varios partidos</h2>
      <p class="footer-note">Sube un archivo .xlsx con una hoja llamada <strong>Partidos</strong> o <strong>Entrada</strong>. Go leerá cada fila, simulará cada partido y generará un Excel de salida con resultados, top marcadores y resumen.</p>
      <div class="actions" style="margin: 12px 0 0;">
        <a class="download-link" href="/download-batch-template">Descargar plantilla de ejemplo</a>
        <span class="hint">Te deja un Excel listo para llenar varios partidos.</span>
      </div>
      <form method="post" action="/process-batch" enctype="multipart/form-data">
        <div class="form-grid" style="margin-top: 14px;">
          <div class="field full">
            <label>Archivo Excel</label>
            <input name="batch_file" type="file" accept=".xlsx">
            <span class="field-help">La hoja debe tener columnas como partido, grupo, equipo_a, equipo_b, tiros_a, tiros_b, motivacion_a, motivacion_b, simulaciones, seed y tiebreaker.</span>
          </div>
          <div class="field">
            <label>Ruta lambda</label>
            <input name="batch_lambda_path" value="{{.BatchLambdaPath}}">
          </div>
          <div class="field">
            <label>Seed global</label>
            <input name="batch_seed" value="{{.BatchSeed}}" placeholder="Opcional">
          </div>
        </div>
        <div class="actions">
          <button type="submit">Procesar Excel</button>
          <span class="hint">Cada fila puede traer hasta 10000 simulaciones.</span>
        </div>
      </form>
    </div>

    <div class="grid">
      <div class="card">
        <h2>Nuevo partido</h2>
        <form method="post" action="/simulate">
          <div class="form-grid">
            <div class="field">
              <label>Ruta lambda</label>
              <input name="lambda_path" value="{{.LambdaPath}}">
            </div>
            <div class="field">
              <label>Semilla</label>
              <input name="seed" value="{{.Seed}}" placeholder="Opcional">
            </div>
            <div class="field">
              <label>Equipo A</label>
              <input name="team_a" value="{{.TeamA}}">
            </div>
            <div class="field">
              <label>Equipo B</label>
              <input name="team_b" value="{{.TeamB}}">
            </div>
            <div class="field">
              <label>Tiros A</label>
              <input name="shots_a" type="number" min="1" value="{{.ShotsA}}">
            </div>
            <div class="field">
              <label>Tiros B</label>
              <input name="shots_b" type="number" min="1" value="{{.ShotsB}}">
            </div>
            <div class="field">
              <label>Motivacion A</label>
              <input name="motivation_a" type="number" min="0" max="10" step="1" value="{{.MotivationA}}" placeholder="0-10" title="0 = nada motivado, 10 = maxima motivacion">
              <span class="field-help">0 = nada motivado, 10 = maxima motivacion.</span>
            </div>
            <div class="field">
              <label>Motivacion B</label>
              <input name="motivation_b" type="number" min="0" max="10" step="1" value="{{.MotivationB}}" placeholder="0-10" title="0 = nada motivado, 10 = maxima motivacion">
              <span class="field-help">0 = nada motivado, 10 = maxima motivacion.</span>
            </div>
            <div class="field">
              <label>Desempate</label>
              <select name="tiebreaker">
                <option value="penalties" {{if eq .Tiebreaker "penalties"}}selected{{end}}>penalties</option>
                <option value="random" {{if eq .Tiebreaker "random"}}selected{{end}}>random</option>
              </select>
            </div>
            <div class="field">
              <label>Simulaciones</label>
              <input name="simulations" type="number" min="1" value="{{.Simulations}}">
            </div>
          </div>
          <div class="actions">
            <button type="submit">Simular y guardar</button>
            <span class="hint">Ejemplo: 6000 corridas para un resumen estable.</span>
          </div>
        </form>
      </div>

      <div class="card">
        <h2>Resultado actual</h2>
        {{if .Result}}
        {{with .Result}}
        {{$report := .}}
        <div class="summary-pill">
          <strong>{{.Run.TeamA}}</strong> vs <strong>{{.Run.TeamB}}</strong>
          <span>|</span>
          <span>{{.Simulations}} simulaciones</span>
        </div>

        <div class="section">
          <h3>Contexto</h3>
          <div class="result-grid">
            <div class="metric"><span>Equipo A</span><strong>{{.Run.TeamA}}</strong><small>{{.Run.ShotsA}} tiros | {{.Run.MotivationA}}/10</small></div>
            <div class="metric"><span>Equipo B</span><strong>{{.Run.TeamB}}</strong><small>{{.Run.ShotsB}} tiros | {{.Run.MotivationB}}/10</small></div>
            <div class="metric"><span>Simulaciones</span><strong>{{.Simulations}}</strong><small>Seed {{.Run.Seed}}</small></div>
            <div class="metric"><span>Desempate</span><strong>{{.Run.Tiebreaker}}</strong><small>Lambda {{.Run.LambdaPath}}</small></div>
          </div>
        </div>

        <div class="section">
          <h3>Lectura rapida</h3>
          <div class="result-grid">
            <div class="metric"><span>Favorito estadistico</span><strong>{{.FavoriteTeam}}</strong><small>{{printf "%.2f%%" .FavoritePct}} de peso</small></div>
            <div class="metric"><span>Ventaja</span><strong>{{printf "%.2f pp" .FavoriteAdvantage}}</strong><small>sobre el rival</small></div>
            <div class="metric"><span>Marcador mas repetido</span><strong>{{.MostRepeatedLabel}}</strong><small>{{printf "%.2f%%" .MostRepeatedPct}} del total</small></div>
            <div class="metric"><span>Promedio total de goles</span><strong>{{printf "%.2f" .AvgTotalGoals}}</strong><small>Diferencia media {{printf "%.2f" .AvgGoalDiff}}</small></div>
          </div>
        </div>

        <div class="section">
          <h3>Distribucion de resultados</h3>
          <div class="result-grid">
            <div class="metric"><span>Victorias {{.Run.TeamA}}</span><strong>{{.Run.WinsA}}</strong><small>{{printf "%.2f%%" .WinPctA}}</small></div>
            <div class="metric"><span>Victorias {{.Run.TeamB}}</span><strong>{{.Run.WinsB}}</strong><small>{{printf "%.2f%%" .WinPctB}}</small></div>
            <div class="metric"><span>Empates en tiempo regular</span><strong>{{.RegulationDraws}}</strong><small>{{printf "%.2f%%" .DrawPct}}</small></div>
            <div class="metric"><span>Definidos por penales</span><strong>{{.Run.Penalties}}</strong><small>{{printf "%.2f%%" .PenaltyPct}}</small></div>
          </div>
        </div>

        <div class="section">
          <h3>Produccion ofensiva</h3>
          <div class="result-grid">
            <div class="metric"><span>Promedio goles {{.Run.TeamA}}</span><strong>{{printf "%.2f" .AvgGoalsA}}</strong><small>{{.Run.GoalsA}} goles totales</small></div>
            <div class="metric"><span>Promedio goles {{.Run.TeamB}}</span><strong>{{printf "%.2f" .AvgGoalsB}}</strong><small>{{.Run.GoalsB}} goles totales</small></div>
            <div class="metric"><span>Total goles acumulados</span><strong>{{.TotalGoals}}</strong><small>entre ambos equipos</small></div>
            <div class="metric"><span>Diferencia total</span><strong>{{.GoalDiff}}</strong><small>acumulada en la serie</small></div>
          </div>
        </div>

        <div class="section">
          <h3>Marcadores exactos destacados</h3>
          <div class="score-grid">
            {{range .Highlighted}}
            <div class="score-card">
              <h4>{{$report.Run.TeamA}} {{.GoalsA}} - {{.GoalsB}} {{$report.Run.TeamB}}</h4>
              <div class="score-meta">
                <span>{{.Count}} veces</span>
                <span>{{printf "%.2f%%" .Percent}}</span>
              </div>
              <div class="score-bar"><span style="width: {{printf "%.2f%%" .Percent}};"></span></div>
            </div>
            {{end}}
          </div>
        </div>

        <div class="section">
          <h3>Top 10 marcadores exactos</h3>
          <table>
            <thead>
              <tr>
                <th>Marcador</th>
                <th>Veces</th>
                <th>Porcentaje</th>
              </tr>
            </thead>
            <tbody>
              {{range .TopScores}}
              <tr>
                <td>{{$report.Run.TeamA}} {{.GoalsA}} - {{.GoalsB}} {{$report.Run.TeamB}}</td>
                <td>{{.Count}}</td>
                <td>{{printf "%.2f%%" .Percent}}</td>
              </tr>
              {{end}}
            </tbody>
          </table>
        </div>
        {{end}}
        {{else}}
        <p>Aun no hay una simulacion guardada en esta sesion.</p>
        {{end}}
      </div>
    </div>

    <div class="card section">
      <h2>Historial reciente</h2>
      {{if .History}}
      <table>
        <thead>
          <tr>
            <th>Fecha</th>
            <th>Partido</th>
            <th>Simulaciones</th>
            <th>Victorias</th>
            <th>Goles promedio</th>
            <th>Definicion</th>
            <th>Semilla</th>
          </tr>
        </thead>
        <tbody>
          {{range .History}}
          <tr>
            <td>{{.DisplayCreatedAt}}</td>
            <td>{{.TeamA}} vs {{.TeamB}}</td>
            <td>{{.Simulations}}</td>
            <td>{{.WinsA}} - {{.WinsB}}</td>
            <td>{{avg .GoalsA .Simulations}} - {{avg .GoalsB .Simulations}}</td>
            <td>{{.Regulation}} reg / {{.Penalties}} pen / {{.RandomTie}} sort</td>
            <td>{{.Seed}}</td>
          </tr>
          {{end}}
        </tbody>
      </table>
      {{else}}
      <p>No hay partidas guardadas todavia.</p>
      {{end}}
    </div>
  </div>
</body>
</html>`
