#!/usr/bin/env perl

use strict;
use warnings;
use Mojo::JSON qw/encode_json decode_json/;
use Mojo::File;

open my $fh, "<", "version/version.go" || die $!;

my $version;
while (<$fh>) {
  $version = $1 if /^const\s+CurrentVersion.*?"v([\d\.]+)"/;
}
close $fh;
die "no version?" unless defined $version;

my @version_parts = split /\./, $version;
die "bad version?" unless defined $version_parts[2];

foreach (@version_parts) {
  $_ = 0 + $_;
}

$version = "v$version";

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
  print "building for $type\n";
  local $ENV{GOOS}   = $build{$type}->{env}->{GOOS};
  local $ENV{GOARCH} = $build{$type}->{env}->{GOARCH};

  unlink "resource.syso";

  my @ldflags = ();
  if ($type eq "win") {
    # create the versioninfo.json based on the current version
    my $tmp = Mojo::File->new("versioninfo.json-template")->slurp();
    my $vdata = decode_json($tmp);
    $vdata->{FixedFileInfo}->{FileVersion}->{Major} = $version_parts[0] ;
    $vdata->{FixedFileInfo}->{FileVersion}->{Minor} = $version_parts[1] ;
    $vdata->{FixedFileInfo}->{FileVersion}->{Patch} = $version_parts[2] ;
    $vdata->{FixedFileInfo}->{ProductVersion}->{Major} = $version_parts[0] ;
    $vdata->{FixedFileInfo}->{ProductVersion}->{Minor} = $version_parts[1] ;
    $vdata->{FixedFileInfo}->{ProductVersion}->{Patch} = $version_parts[2] ;

    $vdata->{StringFileInfo}->{ProductVersion} = $version;

    Mojo::File->new("versioninfo.json")->spurt(encode_json($vdata));

    @ldflags = (qw/ -ldflags -H=windowsgui/);
    system "go", "generate";
  }
  warn join(' ', "go", "build", @ldflags, "-o", "release/$type/" . $build{$type}->{filename});
  system "go", "build", @ldflags, "-o", "release/$type/" . $build{$type}->{filename};
  system "zip", "-j", "dist/dau-$type-$version.zip", ( glob "release/$type/*" );
}

sub add_extras {
  # we used to have a .bat file here, but no longer needed
}
