#!/usr/bin/env perl

use strict;
use warnings;

open my $fh, "<", "dau.go" || die $!;

my $version;
while (<$fh>) {
  $version = $1 if /^const\s+currentVersion.*?"([\d\.]+)"/;
}
close $fh;

die "no version?" unless defined $version;

# so lazy
system "rm", "-rf", "release", "dist";
system "mkdir", "release";
system "mkdir", "dist";

my %build = (
  win   => { env => { GOOS => 'windows', GOARCH => '386' }, filename => 'dau.exe' },
  linux => { env => { GOOS => 'linux',   GOARCH => '386' }, filename => 'dau' },
  mac   => { env => { GOOS => 'darwin',  GOARCH => '386' }, filename => 'dau' },
); 

foreach my $type (keys %build) {
  mkdir "release/$type";
}

add_extras();

system(qw{go-bindata -pkg asset -o asset/asset.go -prefix data/ data});
foreach my $type (keys %build) {
  local $ENV{GOOS}   = $build{$type}->{env}->{GOOS};
  local $ENV{GOARCH} = $build{$type}->{env}->{GOARCH};
  system "go", "build", "-o", "release/$type/" . $build{$type}->{filename};
  system "zip", "-j", "dist/dau-$type-$version.zip", ( glob "release/$type/*" );
}

sub add_extras {
  # bat file for windows

  open (my $fh, ">", "release/win/dau.bat") || die $!;
  print $fh 'set WEBHOOK_URL=https://yourdiscordwebhookURLhere' . "\r\n";
  print $fh 'set SCREENSHOTS="C:\your\screenshot\directory\here"' ."\r\n";
  print $fh 'set USERNAME="Posted by Joe Bloggs"' . "\r\n";
  print $fh 'set WATCH=10' . "\r\n";

  print $fh 'dau.exe --webhook %WEBHOOK_URL% --directory %SCREENSHOTS% --username %USERNAME% --watch %WATCH%' . "\r\n";
  print $fh 'pause' . "\r\n";
  close $fh;
}
