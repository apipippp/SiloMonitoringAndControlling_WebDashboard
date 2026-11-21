@extends('layouts.shell')

@section('title', 'Scan barcode · Silo SCM')

@section('page')
<header class="topbar">
    <div>
        <p class="muted">Tracking batch masuk</p>
        <h1>Scan barcode</h1>
    </div>
    <span class="pill" id="scan-status">Kamera mati</span>
</header>

<div class="scan-grid">
    <section class="card">
        <div class="cam">
            <video id="cam" playsinline muted></video>
            <div class="frame" aria-hidden="true"></div>
        </div>
        <div class="scan-actions">
            <button type="button" class="btn btn-primary" id="btn-cam">Aktifkan kamera</button>
        </div>
        <details class="manual">
            <summary>Tempel data manual</summary>
            <textarea id="manual" rows="5" placeholder='{"id":"BOX-2611-002", ...}'></textarea>
            <button type="button" class="btn" id="btn-manual">Simpan</button>
        </details>
        <p class="hint">Kamera hanya bisa dipakai di localhost atau HTTPS. Kalau dibuka dari HP lewat IP lokal, pakai HTTPS atau input manual.</p>
    </section>

    <section class="card">
        <div class="card-head"><div><span class="kicker">Hasil terakhir</span><h2 id="res-title">Belum ada hasil</h2></div></div>
        <dl class="fields" id="res-fields"></dl>
        <a class="btn btn-primary" id="res-link" hidden>Lihat detail lengkap</a>
        <div class="ctl-msg" id="scan-msg" role="status" aria-live="polite"></div>
    </section>
</div>
@endsection

@push('scripts')
<script src="https://cdn.jsdelivr.net/npm/jsqr@1.4.0/dist/jsQR.js"></script>
<script>window.TRACK = { simpan: @json(route('tracking.store')), csrf: @json(csrf_token()) };</script>
<script src="{{ asset('js/scan.js') }}"></script>
@endpush
