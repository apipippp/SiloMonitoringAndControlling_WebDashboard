(function () {
  const $ = (id) => document.getElementById(id);
  const T = window.TRACK;
  const LABEL = { lot: 'Lot', frm: 'Kode FRM', var: 'Varietas', prc: 'Proses', dt: 'Tanggal', mc: 'Kadar air (%)', grd: 'Grade', wt: 'Berat (kg)', silo: 'Silo tujuan', geo: 'Lokasi' };
  const video = $('cam');
  const kanvas = document.createElement('canvas');

  let stream = null, detektor = null, berjalan = false, terakhir = { kode: null, waktu: 0 };

  function pesan(teks, kelas) {
    const m = $('scan-msg');
    m.textContent = teks;
    m.className = 'ctl-msg' + (kelas ? ' ' + kelas : '');
  }

  function parse(teks) {
    try {
      const d = JSON.parse(teks);
      return d && typeof d === 'object' && d.id ? d : null;
    } catch (e) {
      return null;
    }
  }

  function tampil(data, url) {
    $('res-title').textContent = data.id;
    const wadah = $('res-fields');
    wadah.replaceChildren();
    Object.keys(LABEL).forEach((k) => {
      if (data[k] === undefined) return;
      const div = document.createElement('div');
      const dt = document.createElement('dt');
      const dd = document.createElement('dd');
      dt.textContent = LABEL[k];
      dd.textContent = String(data[k]);
      div.append(dt, dd);
      wadah.append(div);
    });
    const link = $('res-link');
    link.href = url;
    link.hidden = false;
  }

  async function simpan(data) {
    // Cegah kode yang sama tersimpan berulang saat kamera masih mengarah ke barcode yang sama.
    if (data.id === terakhir.kode && Date.now() - terakhir.waktu < 4000) return;
    terakhir = { kode: data.id, waktu: Date.now() };
    try {
      const res = await fetch(T.simpan, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Accept: 'application/json', 'X-CSRF-TOKEN': T.csrf },
        body: JSON.stringify(data),
      });
      const j = await res.json();
      if (!res.ok) throw new Error(j.message || 'Gagal menyimpan');
      tampil(data, j.url);
      pesan('Tersimpan: ' + data.id, 'ok');
    } catch (e) {
      pesan(e.message, 'err');
    }
  }

  async function baca() {
    if (detektor) {
      const hasil = await detektor.detect(video);
      return hasil.length ? hasil[0].rawValue : null;
    }
    if (!window.jsQR || video.videoWidth === 0) return null;
    kanvas.width = video.videoWidth;
    kanvas.height = video.videoHeight;
    const ctx = kanvas.getContext('2d', { willReadFrequently: true });
    ctx.drawImage(video, 0, 0);
    const img = ctx.getImageData(0, 0, kanvas.width, kanvas.height);
    const q = window.jsQR(img.data, img.width, img.height);
    return q ? q.data : null;
  }

  async function loop() {
    while (berjalan) {
      try {
        const teks = await baca();
        if (teks) {
          const data = parse(teks);
          if (data) await simpan(data);
        }
      } catch (e) { /* frame gagal dibaca, coba lagi */ }
      await new Promise((r) => setTimeout(r, 400));
    }
  }

  async function mulai() {
    try {
      stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' }, audio: false });
    } catch (e) {
      pesan('Kamera tidak bisa diakses (' + e.name + '). Pakai input manual.', 'err');
      return;
    }
    video.srcObject = stream;
    await video.play();

    if ('BarcodeDetector' in window) {
      detektor = new BarcodeDetector({ formats: ['qr_code', 'code_128', 'code_39', 'ean_13'] });
    } else if (!window.jsQR) {
      pesan('Pemindai tidak tersedia di peramban ini. Pakai input manual.', 'err');
    }

    berjalan = true;
    $('btn-cam').textContent = 'Matikan kamera';
    $('scan-status').textContent = 'Memindai…';
    loop();
  }

  function henti() {
    berjalan = false;
    if (stream) stream.getTracks().forEach((t) => t.stop());
    stream = null;
    video.srcObject = null;
    $('btn-cam').textContent = 'Aktifkan kamera';
    $('scan-status').textContent = 'Kamera mati';
  }

  $('btn-cam').addEventListener('click', () => (berjalan ? henti() : mulai()));

  $('btn-manual').addEventListener('click', () => {
    const data = parse($('manual').value.trim());
    if (!data) return pesan('Data harus JSON dengan field "id".', 'err');
    simpan(data);
  });
})();
