<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\Schema;

// Data tracking sekarang disimpan di backend Go (dari perangkat), bukan di Laravel.
return new class extends Migration
{
    public function up(): void
    {
        Schema::dropIfExists('tracking_scans');
    }

    public function down(): void
    {
        // tidak dikembalikan: data tracking ada di backend Go
    }
};
