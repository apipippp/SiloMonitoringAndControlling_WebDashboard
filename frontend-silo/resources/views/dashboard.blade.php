@extends('layouts.shell')

@section('title', 'Ringkasan · Silo SCM')

@section('page')
<header class="topbar">
    <div>
        <p class="muted">Silo 01 · Raw material monitoring</p>
        <h1>Halo, Operator</h1>
    </div>
    <div class="top-tools">
        <a href="/cam/" class="btn btn-primary js-cam-link" style="padding:8px 14px;border-radius:999px;font-size:13px;"><svg class="icon" style="width:16px;height:16px;"><use href="#ic-scan"/></svg>Lihat Kamera</a>
        <span class="pill"><span class="dot" id="conn-dot"></span><span id="conn-text">Menghubungkan…</span></span>
        <span class="pill" id="jam-bar">--:--</span>
    </div>
</header>

<div class="dash-main">
        <section class="grid">
            <article class="card hero" id="ringkasan">
                <div class="hero-top">
                    <span class="kicker">Status lingkungan</span>
                    <span class="badge risk-unknown" id="v-risk">-</span>
                </div>
                <div class="hero-main"><span class="big" id="v-suhu">--</span><span class="unit">°C</span></div>
                <p class="muted" id="updated">Menunggu pembacaan</p>
                <div class="hero-row">
                    <div><span class="kicker">Kelembapan</span><strong id="v-hum">--</strong><small>%RH</small></div>
                    <div><span class="kicker">Gas</span><strong id="v-gas">--</strong><small>ADC</small></div>
                    <div><span class="kicker">Jarak</span><strong id="v-jarak">--</strong><small>cm</small></div>
                </div>
            </article>

            <article class="card span-2" id="sensor">
                <div class="card-head">
                    <div><span class="kicker">Tren lingkungan</span><h2>Suhu &amp; kelembapan</h2></div>
                    <select class="sel" id="rentang" aria-label="Rentang waktu">
                        <option value="1">1 jam</option>
                        <option value="6" selected>6 jam</option>
                        <option value="24">24 jam</option>
                        <option value="72">3 hari</option>
                    </select>
                </div>
                <div class="chart-box"><canvas id="chart-env" aria-label="Grafik suhu dan kelembapan"></canvas></div>
            </article>

            <article class="card span-2">
                <div class="card-head">
                    <div><span class="kicker">Gas</span><h2>Pembacaan indikator gas</h2></div>
                    <span class="muted">ADC</span>
                </div>
                <div class="chart-box"><canvas id="chart-gas" aria-label="Grafik gas dan batas kritis"></canvas></div>
            </article>

            <article class="card">
                <div class="card-head"><div><span class="kicker">Level gas</span><h2>Terhadap batas kritis</h2></div></div>
                <div class="ring-wrap">
                    <svg class="ring" viewBox="0 0 180 180" aria-hidden="true">
                        <circle class="track" cx="90" cy="90" r="70"/>
                        <circle class="val" id="ring-val" cx="90" cy="90" r="70" stroke-dasharray="439.82" stroke-dashoffset="439.82"/>
                    </svg>
                    <div class="ring-text"><strong id="ring-pct">--%</strong><span class="muted" id="ring-batas">Menunggu data</span></div>
                </div>
            </article>
        </section>

        <section class="grid-2">
            <article class="card" id="batch">
                <div class="card-head">
                    <div><span class="kicker">Traceability</span><h2>Batch terbaru</h2></div>
                    <span class="pill"><span id="v-today">--</span> hari ini · <span id="v-total">--</span> total</span>
                </div>
                <ul class="list" id="lst-batch"><li class="empty">Memuat data…</li></ul>
            </article>

            <article class="card" id="kejadian">
                <div class="card-head"><div><span class="kicker">Log</span><h2>Kejadian terbaru</h2></div></div>
                <ul class="list" id="lst-event"><li class="empty">Memuat data…</li></ul>
            </article>
        </section>
</div>
@endsection

@push('scripts')
<script src="https://cdn.jsdelivr.net/npm/axios@1/dist/axios.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/chart.js@4/dist/chart.umd.min.js"></script>
<script>
window.SILO_CONFIG = {
    apiUrl: @json(config('silo.public_api_url')),
    token: @json($token),
    segarkanUrl: @json(route('silo.token')),
    csrf: @json(csrf_token()),
};
</script>
<script src="{{ asset('js/silo.js') }}?v=3"></script>
<script src="{{ asset('js/dashboard.js') }}?v=3"></script>
@endpush
