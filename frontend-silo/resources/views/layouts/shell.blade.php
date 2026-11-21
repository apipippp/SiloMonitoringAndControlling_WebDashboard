@extends('layouts.base')

@section('content')
@php
    $menu = [
        ['url' => url('/'), 'ikon' => 'ic-grid', 'label' => 'Ringkasan', 'short' => 'Ringkasan', 'aktif' => 'dashboard', 'hanya_hp' => false],
        ['url' => route('tracking.index'), 'ikon' => 'ic-package', 'label' => 'Riwayat tracking', 'short' => 'Riwayat', 'aktif' => 'tracking.*', 'hanya_hp' => false],
        ['url' => url('/') . '#kendali', 'ikon' => 'ic-sliders', 'label' => 'Kendali perangkat', 'short' => 'Kendali', 'aktif' => null, 'hanya_hp' => true],
    ];
@endphp
<div class="app">
    <aside class="sidebar" aria-label="Menu utama">
        <a href="{{ url('/') }}" class="brand">
            <span class="logo"><svg class="icon"><use href="#ic-cpu"/></svg></span>
            <span class="brand-text"><strong>Silo SCM</strong><small>Monitor &amp; tracking</small></span>
        </a>
        <span class="menu-label">Menu</span>
        <nav class="menu">
            @foreach ($menu as $m)
                @continue($m['hanya_hp'])
                <a href="{{ $m['url'] }}" title="{{ $m['label'] }}" class="nav-link {{ $m['aktif'] && request()->routeIs($m['aktif']) ? 'is-active' : '' }}">
                    <svg class="icon"><use href="#{{ $m['ikon'] }}"/></svg><span>{{ $m['label'] }}</span>
                </a>
            @endforeach
        </nav>
        <div class="device-card">
            <svg class="icon"><use href="#ic-box"/></svg>
            <div><strong>Silo 01</strong><small>Raw material</small></div>
        </div>
    </aside>

    <main class="main">@yield('page')</main>

    <nav class="bottom-nav" aria-label="Menu bawah">
        @foreach ($menu as $m)
            <a href="{{ $m['url'] }}" class="{{ $m['aktif'] && request()->routeIs($m['aktif']) ? 'is-active' : '' }}">
                <svg class="icon"><use href="#{{ $m['ikon'] }}"/></svg><span>{{ $m['short'] }}</span>
            </a>
        @endforeach
    </nav>
</div>
@endsection
