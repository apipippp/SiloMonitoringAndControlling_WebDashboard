<?php

namespace App\Http\Controllers;

use Illuminate\Http\JsonResponse;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Http;

class SiloTokenController extends Controller
{
    private const KUNCI = 'silo.api_token';

    // Login ke API Go memakai akun operator dari .env. Token disimpan di cache Laravel,
    // masa simpannya lebih pendek daripada masa berlaku JWT (JWT_HOURS).
    public static function token(bool $segarkan = false): ?string
    {
        if ($segarkan) {
            Cache::forget(self::KUNCI);
        }

        return Cache::remember(self::KUNCI, now()->addHours(6), function (): ?string {
            try {
                $respons = Http::acceptJson()
                    ->timeout(8)
                    ->post(rtrim(config('silo.api_url'), '/') . '/auth/login', [
                        'email' => config('silo.operator_email'),
                        'password' => config('silo.operator_password'),
                    ]);
            } catch (\Throwable) {
                return null;
            }

            return $respons->successful() ? $respons->json('data.token') : null;
        });
    }

    public function segarkan(): JsonResponse
    {
        $token = self::token(segarkan: true);

        return $token
            ? response()->json(['token' => $token])
            : response()->json(['pesan' => 'Gagal login ke API Go. Cek SILO_OPERATOR_* di .env.'], 503);
    }
}
