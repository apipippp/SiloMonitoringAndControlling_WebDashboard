<?php

namespace App\Http\Controllers;

use Illuminate\View\View;

class TrackingController extends Controller
{
    public function index(): View
    {
        return view('tracking.index', ['token' => SiloTokenController::token()]);
    }

    public function show(string $kode): View
    {
        return view('tracking.show', ['kode' => $kode, 'token' => SiloTokenController::token()]);
    }
}
