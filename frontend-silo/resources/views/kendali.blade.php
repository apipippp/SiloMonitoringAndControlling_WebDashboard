@extends('layouts.shell')

@section('title', 'Kendali Perangkat · Silo SCM')

@section('page')
<header class="topbar">
    <div>
        <p class="muted">Silo 01 · Kontrol aktuator &amp; mode operasi</p>
        <h1>Kendali Perangkat</h1>
    </div>
    <div class="top-tools">
        <a href="/cam/" class="btn btn-primary js-cam-link" style="padding:8px 14px;border-radius:999px;font-size:13px;"><svg class="icon" style="width:16px;height:16px;"><use href="#ic-scan"/></svg>Lihat Kamera</a>
        <span class="pill"><span class="dot" id="conn-dot"></span><span id="conn-text">Menghubungkan…</span></span>
    </div>
</header>

<div class="grid-2" style="align-items:start;">
    <section class="card stack" id="kendali" aria-label="Kendali operator" style="gap:18px;">
        <div class="profile">
            <span class="av av-lg">O</span>
            <strong>Operator</strong>
            <span class="muted">Silo 01 · <span id="v-mode">--</span></span>
        </div>

        <div class="stack">
            <span class="kicker">Mode kendali</span>
            <div class="seg">
                <button type="button" id="btn-mode-auto" data-cmd='{"mode":"AUTO"}'>Otomatis</button>
                <button type="button" id="btn-mode-manual" data-cmd='{"mode":"MANUAL"}'>Manual</button>
            </div>
        </div>

        <div class="stack">
            <span class="kicker">Perangkat</span>
            <div class="sw-row">
                <span class="sw-ico"><svg class="icon"><use href="#ic-fan"/></svg></span>
                <div class="grow"><strong>Kipas</strong><small id="s-fan">--</small></div>
                <button type="button" class="switch" id="sw-fan" role="switch" aria-checked="false" aria-label="Kipas"
                        data-on='{"mode":"MANUAL","fan":"ON"}' data-off='{"mode":"MANUAL","fan":"OFF"}'></button>
            </div>
            <div class="sw-row">
                <span class="sw-ico"><svg class="icon"><use href="#ic-wind"/></svg></span>
                <div class="grow"><strong>Ventilasi</strong><small id="s-vent">--</small></div>
                <button type="button" class="switch" id="sw-vent" role="switch" aria-checked="false" aria-label="Ventilasi"
                        data-on='{"mode":"MANUAL","vent":"OPEN"}' data-off='{"mode":"MANUAL","vent":"CLOSE"}'></button>
            </div>
            <div class="sw-row">
                <span class="sw-ico"><svg class="icon"><use href="#ic-box"/></svg></span>
                <div class="grow"><strong>Konveyor</strong><small id="s-conv">--</small></div>
                <button type="button" class="switch" id="sw-conv" role="switch" aria-checked="false" aria-label="Konveyor"
                        data-on='{"mode":"MANUAL","conveyor":"START"}' data-off='{"mode":"MANUAL","conveyor":"STOP"}'></button>
            </div>
        </div>

        <p class="hint">Menyentuh kipas, ventilasi, atau konveyor akan memindahkan mode ke Manual. Klik <strong>Otomatis</strong> untuk mengembalikan kontrol penuh ke sensor ESP32.</p>
        <div class="ctl-msg" id="ctl-msg" role="status" aria-live="polite"></div>
    </section>

    <section class="card hero">
        <div class="hero-top">
            <span class="kicker">Pembacaan sensor langsung</span>
            <span class="badge risk-unknown" id="v-risk">-</span>
        </div>
        <div class="hero-main"><span class="big" id="v-suhu">--</span><span class="unit">°C</span></div>
        <p class="muted" id="updated">Menunggu pembacaan</p>
        <div class="hero-row">
            <div><span class="kicker">Kelembapan</span><strong id="v-hum">--</strong><small>%RH</small></div>
            <div><span class="kicker">Gas</span><strong id="v-gas">--</strong><small>ADC</small></div>
            <div><span class="kicker">Jarak</span><strong id="v-jarak">--</strong><small>cm</small></div>
        </div>
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
<script src="{{ asset('js/silo.js') }}?v=3"></script>
<script src="{{ asset('js/kendali.js') }}?v=1"></script>
@endpush
