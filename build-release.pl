#!/usr/bin/env perl

use strict;
use warnings;

open my $fh, "<", "version/version.go" || die $!;

my $version;
while (<$fh>) {
  $version = $1 if /^const\s+CurrentVersion.*?"(v[\d\.]+)"/;
}
close $fh;

die "no version?" unless defined $version;

# quit if tests fail
system("go test ./...") && die "not building release with failing tests";

# so lazy
system "rm", "-rf", "release", "dist";
system "mkdir", "release";
system "mkdir", "dist";

my %build = (
  win   => { env => { GOOS => 'windows', GOARCH => 'amd64' }, filename => 'dau.exe' },
  linux => { env => { GOOS => 'linux',   GOARCH => 'amd64' }, filename => 'dau' },
  mac   => { env => { GOOS => 'darwin',  GOARCH => 'amd64' }, filename => 'dau' },
); 

foreach my $type (keys %build) {
  mkdir "release/$type";
}

add_extras();

foreach my $type (keys %build) {
  local $ENV{GOOS}   = $build{$type}->{env}->{GOOS};
  local $ENV{GOARCH} = $build{$type}->{env}->{GOARCH};
  system "go", "build", "-o", "release/$type/" . $build{$type}->{filename};
  system "zip", "-j", "dist/dau-$type-$version.zip", ( glob "release/$type/*" );
}

sub add_extras {
  # we used to have a .bat file here, but no longer needed
}
