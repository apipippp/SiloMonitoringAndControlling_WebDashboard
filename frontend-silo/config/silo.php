<?php

return [
    // Dipakai Laravel (server-side) untuk login ke API Go.
    'api_url' => env('SILO_API_URL', 'http://127.0.0.1:3000/api'),

    // Dipakai JavaScript di peramban untuk memanggil API Go ('auto' -> otomatis ikut host:3000).
    'public_api_url' => env('SILO_PUBLIC_API_URL', 'auto'),

    // Akun operator untuk login server-side. Tidak pernah dikirim ke peramban.
    'operator_email' => env('SILO_OPERATOR_EMAIL', 'admin@silo.local'),
    'operator_password' => env('SILO_OPERATOR_PASSWORD', 'admin123'),
];
