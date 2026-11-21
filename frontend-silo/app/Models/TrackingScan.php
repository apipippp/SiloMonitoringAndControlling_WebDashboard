<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class TrackingScan extends Model
{
    protected $fillable = ['kode', 'silo', 'payload', 'dipindai_pada'];

    protected $casts = [
        'payload' => 'array',
        'dipindai_pada' => 'datetime',
    ];

    // Label Indonesia untuk field JSON barcode (urutan tampil).
    public const LABEL = [
        'lot' => 'Lot',
        'frm' => 'Kode FRM',
        'var' => 'Varietas',
        'prc' => 'Proses',
        'dt' => 'Tanggal',
        'mc' => 'Kadar air (%)',
        'grd' => 'Grade',
        'wt' => 'Berat (kg)',
        'silo' => 'Silo tujuan',
    ];

    // "geo": "-2.9612,119.8981|1680m" -> lat, lng, ketinggian
    public function geo(): ?array
    {
        $g = $this->payload['geo'] ?? '';

        if (! preg_match('/^(-?\d+(?:\.\d+)?),(-?\d+(?:\.\d+)?)\|(-?\d+)m$/', $g, $m)) {
            return null;
        }

        return ['lat' => $m[1], 'lng' => $m[2], 'alt' => $m[3]];
    }
}
