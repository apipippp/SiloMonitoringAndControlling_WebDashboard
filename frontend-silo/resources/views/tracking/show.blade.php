@extends('layouts.shell')

@section('title', $kode . ' · Silo SCM')

@section('page')
<header class="topbar">
    <div>
        <a class="back" href="{{ route('tracking.index') }}"><svg class="icon"><use href="#ic-arrow-left"/></svg> Riwayat</a>
        <h1 id="det-title">{{ $kode }}</h1>
        <p class="muted" id="det-sub">Memuat data…</p>
    </div>
    <span class="pill" id="det-time">—</span>
</header>

<div id="detail-root" data-kode="{{ $kode }}" class="stack-main">
    <section class="card">
        <div class="card-head"><div><span class="kicker">Data batch</span><h2>Detail lengkap</h2></div></div>
        <dl class="fields" id="det-fields"></dl>
        <p class="muted" id="det-status"></p>
    </section>

    <section class="card" id="det-snap-card" style="display:none">
        <div class="card-head"><div><span class="kicker">Tangkapan kamera</span><h2>Screenshot saat barang VALID</h2></div></div>
        <div style="text-align:center">
            <a id="det-snap-link" href="#" target="_blank" rel="noopener">
                <img id="det-snap-img" alt="Snapshot Kamera Barang Valid" style="max-width:100%;max-height:420px;border-radius:12px;border:1px solid #e5d8f7;box-shadow:0 6px 20px rgba(0,0,0,0.08)">
            </a>
        </div>
    </section>

    <section class="card">
        <div class="card-head"><div><span class="kicker">Data mentah</span><h2>JSON dari perangkat</h2></div></div>
        <pre class="raw" id="det-raw"></pre>
    </section>
</div>
@endsection

@push('scripts')
<script src="https://cdn.jsdelivr.net/npm/axios@1/dist/axios.min.js"></script>
<script>
window.SILO_CONFIG = {
    apiUrl: @json(config('silo.public_api_url')),
    token: @json($token),
    segarkanUrl: @json(route('silo.token')),
    csrf: @json(csrf_token()),
};
</script>
<script src="{{ asset('js/silo.js') }}"></script>
<script src="{{ asset('js/tracking.js') }}"></script>
@endpush
