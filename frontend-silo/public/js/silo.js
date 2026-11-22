// Klien API bersama. Token datang dari server Laravel; jika ditolak (401) diminta ulang sekali lalu permintaan diulang.
window.Silo = (function () {
  const cfg = window.SILO_CONFIG;
  const isDirectLaravel = window.location.port === '8000';
  const resolvedApiUrl = (!cfg.apiUrl || cfg.apiUrl === 'auto')
    ? (isDirectLaravel
        ? (window.location.protocol + '//' + window.location.hostname + ':3000/api')
        : (window.location.origin + '/api'))
    : cfg.apiUrl;
  const apiOrigin = resolvedApiUrl.replace(/\/api\/?$/, '');
  cfg.apiUrl = resolvedApiUrl;

  const api = axios.create({ baseURL: resolvedApiUrl, timeout: 10000, headers: { Accept: 'application/json' } });

  api.interceptors.request.use((k) => {
    k.headers.Authorization = 'Bearer ' + cfg.token;
    return k;
  });

  let segarkan = null; // satu permintaan token baru untuk semua panggilan yang gagal bersamaan
  function tokenBaru() {
    if (!segarkan) {
      segarkan = fetch(cfg.segarkanUrl, { method: 'POST', headers: { 'X-CSRF-TOKEN': cfg.csrf, Accept: 'application/json' } })
        .then((r) => (r.ok ? r.json() : Promise.reject(new Error('token'))))
        .then((j) => { cfg.token = j.token; })
        .finally(() => { segarkan = null; });
    }
    return segarkan;
  }

  api.interceptors.response.use(
    (r) => r,
    async (galat) => {
      const asli = galat.config;
      if (galat.response && galat.response.status === 401 && asli && !asli._ulang) {
        asli._ulang = true;
        try { await tokenBaru(); return api.request(asli); } catch (e) { /* lanjut ke galat asli */ }
      }
      return Promise.reject(galat);
    }
  );

  const pad = (n) => String(n).padStart(2, '0');

  function fmt(v, d) {
    if (v === null || v === undefined || Number.isNaN(Number(v))) return '--';
    return Number(v).toFixed(d === undefined ? 1 : d);
  }
  function jam(ts) { const d = new Date(ts * 1000); return pad(d.getHours()) + ':' + pad(d.getMinutes()); }
  function jamDetik(ts) { const d = new Date(ts * 1000); return jam(ts) + ':' + pad(d.getSeconds()); }
  function tanggalJam(ts) { const d = new Date(ts * 1000); return pad(d.getDate()) + '/' + pad(d.getMonth() + 1) + ' ' + jam(ts); }

  // Selalu textContent (BUKAN innerHTML): data dari perangkat tidak dipercaya.
  function el(tag, text, cls) {
    const e = document.createElement(tag);
    if (text !== undefined && text !== null) e.textContent = text;
    if (cls) e.className = cls;
    return e;
  }

  function pesanGalat(galat) {
    return (galat.response && galat.response.data && galat.response.data.pesan) || 'Tidak dapat menghubungi server API';
  }

  function riskClass(level) {
    switch (String(level || '').toUpperCase()) {
      case 'NORMAL': return 'risk-normal';
      case 'WARNING': return 'risk-warning';
      case 'CRITICAL': return 'risk-critical';
      default: return 'risk-unknown';
    }
  }

  const gaya = { suhu: '#8b5cf6', hum: '#f472b6', gas: '#fbbf24', crit: '#f43f5e', grid: '#f0e7fb', teks: '#8c7aa8' };

  if (window.Chart) {
    Chart.defaults.color = gaya.teks;
    Chart.defaults.font.family = '"Plus Jakarta Sans", system-ui, sans-serif';
    Chart.defaults.font.size = 11;
    Chart.defaults.borderColor = gaya.grid;
  }

  return { cfg, api, apiOrigin, fmt, jam, jamDetik, tanggalJam, el, pesanGalat, riskClass, gaya };
})();
