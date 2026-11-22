<?php

namespace App\Http\Controllers;

use Illuminate\View\View;

class DashboardController extends Controller
{
    public function dashboard(): View
    {
        return view('dashboard', ['token' => SiloTokenController::token()]);
    }

    public function kendali(): View
    {
        return view('kendali', ['token' => SiloTokenController::token()]);
    }
}
