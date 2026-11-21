(function () {
  const S = window.Silo;
  const $ = (id) => document.getElementById(id);
  const KELILING = 2 * Math.PI * 70; // keliling lingkaran ring (r = 70)
  let envChart, gasChart;

  function jamBar() {
    const hari = new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'short' });
    $('jam-bar').textContent = hari + ' · ' + S.jam(Date.now() / 1000);
  }

  function setKoneksi(online, teks) {
    $('conn-dot').className = 'dot ' + (online ? 'on' : 'off');
    $('conn-text').textContent = teks;
  }

  function setSwitch(id, aktif) {
    $(id).setAttribute('aria-checked', aktif ? 'true' : 'false');
  }

  function render(d) {
    const r = d.pembacaan;
    const b = d.batas;
    $('v-total').textContent = d.barang.total;
    $('v-today').textContent = d.barang.hari_ini;

    if (!r) {
      setKoneksi(false, 'Belum ada data');
      return;
    }

    setKoneksi(d.online, d.online ? 'Perangkat online' : 'Offline · ' + d.umur_detik + ' dtk');
    $('updated').textContent = 'Data terakhir ' + S.jamDetik(r.ts);
    $('v-suhu').textContent = S.fmt(r.suhu, 1);
    $('v-hum').textContent = S.fmt(r.kelembaban, 0);
    $('v-gas').textContent = S.fmt(r.gas, 0);
    $('v-jarak').textContent = S.fmt(r.jarak_cm, 0);

    const badge = $('v-risk');
    badge.className = 'badge ' + S.riskClass(r.risiko);
    badge.textContent = r.risiko;

    const pct = Math.min(100, Math.max(0, (r.gas / b.crit_gas) * 100));
    $('ring-val').style.strokeDashoffset = KELILING * (1 - pct / 100);
    $('ring-pct').textContent = Math.round(pct) + '%';
    $('ring-batas').textContent = 'dari batas CRITICAL ' + b.crit_gas;

    // Kendali
    const manual = r.mode === 'MANUAL';
    $('v-mode').textContent = 'mode ' + String(r.mode).toLowerCase();
    $('btn-mode-auto').classList.toggle('is-active', !manual);
    $('btn-mode-manual').classList.toggle('is-active', manual);
    setSwitch('sw-fan', r.kipas);
    setSwitch('sw-vent', r.ventilasi);
    setSwitch('sw-conv', String(r.konveyor).toUpperCase() === 'RUNNING');
    $('s-fan').textContent = (r.kipas ? 'Menyala' : 'Mati') + ' · ' + (manual ? 'manual' : 'otomatis');
    $('s-vent').textContent = r.ventilasi ? 'Terbuka' : 'Tertutup';
    $('s-conv').textContent = String(r.konveyor).toLowerCase();
  }

  function buatGrafik() {
    const garis = (label, warna, sumbuY) => ({
      label, data: [], borderColor: warna, backgroundColor: warna,
      borderWidth: 2.5, pointRadius: 0, tension: 0.35, yAxisID: sumbuY,
    });
    const umum = {
      responsive: true,
      maintainAspectRatio: false,
      interaction: { mode: 'index', intersect: false },
      plugins: { legend: { position: 'bottom', labels: { usePointStyle: true, boxWidth: 8 } } },
    };

    envChart = new Chart($('chart-env'), {
      type: 'line',
      data: { labels: [], datasets: [garis('Suhu (°C)', S.gaya.suhu, 'y'), garis('Kelembapan (%RH)', S.gaya.hum, 'y1')] },
      options: Object.assign({}, umum, {
        scales: {
          x: { grid: { display: false }, ticks: { maxTicksLimit: 8 } },
          y: { position: 'left', grid: { color: S.gaya.grid } },
          y1: { position: 'right', grid: { display: false } },
        },
      }),
    });

    gasChart = new Chart($('chart-gas'), {
      type: 'bar',
      data: {
        labels: [],
        datasets: [
          { label: 'Gas (ADC)', data: [], backgroundColor: S.gaya.gas, borderRadius: 6, maxBarThickness: 18 },
          { label: 'Batas CRITICAL', type: 'line', data: [], borderColor: S.gaya.crit, borderDash: [6, 4], borderWidth: 1.5, pointRadius: 0 },
        ],
      },
      options: Object.assign({}, umum, {
        scales: { x: { grid: { display: false }, ticks: { maxTicksLimit: 8 } }, y: { grid: { color: S.gaya.grid } } },
      }),
    });
  }

  async function muatRiwayat() {
    const jam = Number($('rentang').value);
    try {
      const d = (await S.api.get('/telemetri/riwayat', { params: { jam } })).data.data;
      const label = d.titik.map((p) => (jam > 24 ? S.tanggalJam(p.ts) : S.jam(p.ts)));

      envChart.data.labels = label;
      envChart.data.datasets[0].data = d.titik.map((p) => p.suhu);
      envChart.data.datasets[1].data = d.titik.map((p) => p.kelembaban);
      envChart.update('none');

      gasChart.data.labels = label;
      gasChart.data.datasets[0].data = d.titik.map((p) => p.gas);
      gasChart.data.datasets[1].data = d.titik.map(() => d.batas.crit_gas);
      gasChart.update('none');
    } catch (e) { /* grafik dibiarkan apa adanya */ }
  }

  function isiList(ul, items, kosong) {
    ul.replaceChildren();
    if (items.length === 0) { ul.append(S.el('li', kosong, 'empty')); return; }
    items.forEach((li) => ul.append(li));
  }

  async function muatBatch() {
    try {
      const list = (await S.api.get('/barang', { params: { limit: 6 } })).data.data;
      isiList($('lst-batch'), list.map((b) => {
        const li = S.el('li', null, 'row');
        li.append(S.el('span', (b.material || '?').charAt(0).toUpperCase(), 'av'));
        const main = S.el('div', null, 'row-main');
        main.append(S.el('strong', b.material), S.el('span', b.batch_id, 'mono'));
        li.append(main, S.el('span', S.tanggalJam(b.ts), 'mono'));
        return li;
      }), 'Belum ada barang masuk.');
    } catch (e) { /* abaikan */ }
  }

  async function muatKejadian() {
    try {
      const list = (await S.api.get('/kejadian', { params: { limit: 6 } })).data.data;
      isiList($('lst-event'), list.map((k) => {
        const li = S.el('li', null, 'row');
        li.append(S.el('span', k.level, 'tag tag-' + k.level.toLowerCase()));
        const main = S.el('div', null, 'row-main');
        main.append(S.el('strong', k.pesan), S.el('span', S.tanggalJam(k.ts), 'mono'));
        li.append(main);
        return li;
      }), 'Belum ada kejadian.');
    } catch (e) { /* abaikan */ }
  }

  async function muatTerbaru() {
    try {
      render((await S.api.get('/telemetri/terbaru')).data.data);
    } catch (e) {
      setKoneksi(false, 'API tidak terjangkau');
    }
  }

  function pesan(teks, kelas) {
    const m = $('ctl-msg');
    m.textContent = teks;
    m.className = 'ctl-msg' + (kelas ? ' ' + kelas : '');
  }

  async function kirim(perintah, tombol) {
    tombol.disabled = true;
    pesan('Mengirim perintah…');
    try {
      await S.api.post('/kontrol', perintah);
      pesan('Terkirim: ' + Object.entries(perintah).map(([k, v]) => k + ' → ' + v).join(', '), 'ok');
      setTimeout(muatTerbaru, 1500);
    } catch (e) {
      pesan(S.pesanGalat(e), 'err');
    } finally {
      tombol.disabled = false;
    }
  }

  function pasangKendali() {
    document.querySelectorAll('.seg button[data-cmd]').forEach((b) =>
      b.addEventListener('click', () => kirim(JSON.parse(b.dataset.cmd), b)));

    document.querySelectorAll('.switch').forEach((sw) =>
      sw.addEventListener('click', () => {
        const aktif = sw.getAttribute('aria-checked') === 'true';
        kirim(JSON.parse(aktif ? sw.dataset.off : sw.dataset.on), sw);
      }));
  }

  jamBar();
  setInterval(jamBar, 30000);
  buatGrafik();
  pasangKendali();
  $('rentang').addEventListener('change', muatRiwayat);

  muatTerbaru();
  muatRiwayat();
  muatBatch();
  muatKejadian();

  setInterval(muatTerbaru, 3000);
  setInterval(() => { muatBatch(); muatKejadian(); }, 5000);
  setInterval(muatRiwayat, 30000);
})();
