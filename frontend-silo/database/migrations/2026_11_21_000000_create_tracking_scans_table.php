<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('tracking_scans', function (Blueprint $table) {
            $table->id();
            $table->string('kode', 64)->unique();
            $table->string('silo', 32)->nullable();
            $table->json('payload');
            $table->timestamp('dipindai_pada');
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('tracking_scans');
    }
};
