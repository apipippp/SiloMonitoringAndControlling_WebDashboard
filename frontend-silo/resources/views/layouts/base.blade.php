<!doctype html>
<html lang="id">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="csrf-token" content="{{ csrf_token() }}">
    <meta name="theme-color" content="#a78bfa">
    <title>@yield('title', 'Silo SCM Monitor')</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Bricolage+Grotesque:opsz,wght@12..96,500;600;700&family=Plus+Jakarta+Sans:wght@400;500;600;700&family=IBM+Plex+Mono:wght@400;500&display=swap" rel="stylesheet">
    <link rel="stylesheet" href="{{ asset('css/dashboard.css') }}">
    @stack('styles')
</head>
<body>
<!-- Ikon SVG disimpan secara lokal di halaman; tidak memerlukan paket ikon eksternal. -->
<svg xmlns="http://www.w3.org/2000/svg" class="icon-sprite" aria-hidden="true" focusable="false">
    <symbol id="ic-activity" viewBox="0 0 24 24"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></symbol>
    <symbol id="ic-grid" viewBox="0 0 24 24"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></symbol>
    <symbol id="ic-chart" viewBox="0 0 24 24"><path d="M3 3v18h18"/><path d="m7 16 4-5 4 3 5-8"/></symbol>
    <symbol id="ic-thermo" viewBox="0 0 24 24"><path d="M14 14.76V5a3 3 0 0 0-6 0v9.76a5 5 0 1 0 6 0Z"/><path d="M11 11v7"/></symbol>
    <symbol id="ic-droplet" viewBox="0 0 24 24"><path d="M12 22a8 8 0 0 0 8-8c0-4.4-8-12-8-12S4 9.6 4 14a8 8 0 0 0 8 8Z"/></symbol>
    <symbol id="ic-wind" viewBox="0 0 24 24"><path d="M3 8h12a3 3 0 1 0-3-3"/><path d="M2 12h18a3 3 0 1 1-3 3"/><path d="M4 16h7a3 3 0 1 1-3 3"/></symbol>
    <symbol id="ic-package" viewBox="0 0 24 24"><path d="m12 2 9 5-9 5-9-5 9-5Z"/><path d="M3 7v10l9 5 9-5V7"/><path d="M12 12v10"/></symbol>
    <symbol id="ic-fan" viewBox="0 0 24 24"><path d="M12 12c-1.5-2.5-1.2-8 2-9 2-.6 4 1 4 3 0 2-2.7 4.2-6 6Z"/><path d="M12 12c2.9 0 7.6 2.6 7.4 5.9-.2 2.1-2.5 3.2-4.3 2.3C13.4 19.4 12 16 12 12Z"/><path d="M12 12c-1.4 2.5-6.1 5.3-8.5 3.4-1.6-1.3-1.4-3.8.4-4.7C5.7 9.9 9.1 10.4 12 12Z"/><circle cx="12" cy="12" r="2"/></symbol>
    <symbol id="ic-cpu" viewBox="0 0 24 24"><rect x="5" y="5" width="14" height="14" rx="2"/><path d="M9 9h6v6H9zM9 1v4m6-4v4M9 19v4m6-4v4M1 9h4m-4 6h4m14-6h4m-4 6h4"/></symbol>
    <symbol id="ic-box" viewBox="0 0 24 24"><path d="m3 6 9-4 9 4v12l-9 4-9-4V6Z"/><path d="m3 6 9 4 9-4M12 10v12"/></symbol>
    <symbol id="ic-scan" viewBox="0 0 24 24"><path d="M4 8V5a1 1 0 0 1 1-1h3m8 0h3a1 1 0 0 1 1 1v3m0 8v3a1 1 0 0 1-1 1h-3m-8 0H5a1 1 0 0 1-1-1v-3M4 12h16"/></symbol>
    <symbol id="ic-arrow-right" viewBox="0 0 24 24"><path d="M5 12h14m-6-6 6 6-6 6"/></symbol>
    <symbol id="ic-arrow-left" viewBox="0 0 24 24"><path d="M19 12H5m6 6-6-6 6-6"/></symbol>
    <symbol id="ic-log-out" viewBox="0 0 24 24"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><path d="M16 17l5-5-5-5m5 5H9"/></symbol>
    <symbol id="ic-lock" viewBox="0 0 24 24"><rect x="4" y="11" width="16" height="10" rx="2"/><path d="M8 11V7a4 4 0 0 1 8 0v4"/></symbol>
    <symbol id="ic-database" viewBox="0 0 24 24"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14a9 3 0 0 0 18 0V5M3 12a9 3 0 0 0 18 0"/></symbol>
    <symbol id="ic-sliders" viewBox="0 0 24 24"><path d="M4 7h9m4 0h3M4 17h3m4 0h9"/><circle cx="15" cy="7" r="2"/><circle cx="9" cy="17" r="2"/></symbol>
    <symbol id="ic-clock" viewBox="0 0 24 24"><circle cx="12" cy="12" r="9"/><path d="M12 7v5l4 2"/></symbol>
    <symbol id="ic-info" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M12 11v6m0-10h.01"/></symbol>
    <symbol id="ic-chevron-down" viewBox="0 0 24 24"><path d="m6 9 6 6 6-6"/></symbol>
    <symbol id="ic-chevron-right" viewBox="0 0 24 24"><path d="m9 6 6 6-6 6"/></symbol>
    <symbol id="ic-shield" viewBox="0 0 24 24"><path d="m12 22 7-4V6l-7-4-7 4v12l7 4Z"/><path d="m9 12 2 2 4-4"/></symbol>
    <symbol id="ic-menu" viewBox="0 0 24 24"><path d="M4 7h16M4 12h16M4 17h16"/></symbol>
    <symbol id="ic-power" viewBox="0 0 24 24"><path d="M12 2v10"/><path d="M18.4 6.6a9 9 0 1 1-12.8 0"/></symbol>
    <symbol id="ic-play" viewBox="0 0 24 24"><path d="M6 4l14 8-14 8V4Z"/></symbol>
    <symbol id="ic-stop" viewBox="0 0 24 24"><rect x="5" y="5" width="14" height="14" rx="2"/></symbol>
    <symbol id="ic-auto" viewBox="0 0 24 24"><path d="M21 12a9 9 0 0 1-15.5 6.2M3 12a9 9 0 0 1 15.5-6.2"/><path d="M18 2v4h-4M6 22v-4h4"/></symbol>
    <symbol id="ic-eye" viewBox="0 0 24 24"><path d="M1 12s4-7 11-7 11 7 11 7-4 7-11 7S1 12 1 12Z"/><circle cx="12" cy="12" r="3"/></symbol>
</svg>
@yield('content')
@stack('scripts')
</body>
</html>
