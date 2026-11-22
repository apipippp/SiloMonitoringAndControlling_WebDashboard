(function () {
  const S = window.Silo;
  const $ = (id) => document.getElementById(id);

  function setKoneksi(online, teks) {
    $('conn-dot').className = 'dot ' + (online ? 'on' : 'off');
    $('conn-text').textContent = teks;
  }

  function setSwitch(id, aktif) {
    $(id).setAttribute('aria-checked', aktif ? 'true' : 'false');
  }

  function render(d) {
    const r = d.pembacaan;
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

    const manual = String(r.mode || '').toUpperCase() === 'MANUAL';
    const convUp = String(r.konveyor || '').toUpperCase();
    const convLabel = { RUNNING: 'Berjalan', STOPPED: 'Berhenti', SCANNING: 'Memindai QR', ERROR: 'Tahan / Invalid' }[convUp] || convUp.toLowerCase();
    $('v-mode').textContent = 'mode ' + (manual ? 'manual' : 'otomatis');
    $('btn-mode-auto').classList.toggle('is-active', !manual);
    $('btn-mode-manual').classList.toggle('is-active', manual);
    setSwitch('sw-fan', r.kipas);
    setSwitch('sw-vent', r.ventilasi);
    setSwitch('sw-conv', convUp === 'RUNNING');
    $('s-fan').textContent = (r.kipas ? 'Menyala' : 'Mati') + ' · ' + (manual ? 'manual' : 'otomatis');
    $('s-vent').textContent = r.ventilasi ? 'Terbuka' : 'Tertutup';
    $('s-conv').textContent = convLabel + (manual ? ' · manual' : '');
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
      if (perintah.mode === 'MANUAL') {
        $('btn-mode-auto').classList.remove('is-active');
        $('btn-mode-manual').classList.add('is-active');
        $('v-mode').textContent = 'mode manual';
      } else if (perintah.mode === 'AUTO') {
        $('btn-mode-auto').classList.add('is-active');
        $('btn-mode-manual').classList.remove('is-active');
        $('v-mode').textContent = 'mode otomatis';
      }
      if (perintah.fan) setSwitch('sw-fan', perintah.fan === 'ON');
      if (perintah.vent) setSwitch('sw-vent', perintah.vent === 'OPEN');
      if (perintah.conveyor) setSwitch('sw-conv', perintah.conveyor === 'START');
      pesan('Terkirim: ' + Object.entries(perintah).map(([k, v]) => k + ' → ' + v).join(', '), 'ok');
      setTimeout(muatTerbaru, 900);
    } catch (e) {
      pesan(S.pesanGalat(e), 'err');
    } finally {
      tombol.disabled = false;
    }
  }

  document.querySelectorAll('.seg button[data-cmd]').forEach((b) =>
    b.addEventListener('click', () => kirim(JSON.parse(b.dataset.cmd), b)));

  document.querySelectorAll('.switch').forEach((sw) =>
    sw.addEventListener('click', () => {
      const aktif = sw.getAttribute('aria-checked') === 'true';
      kirim(JSON.parse(aktif ? sw.dataset.off : sw.dataset.on), sw);
    }));

  muatTerbaru();
  setInterval(muatTerbaru, 2500);
})();
