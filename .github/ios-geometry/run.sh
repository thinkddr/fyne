#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2026 Daniel Delgado Ramos
set -euo pipefail
root=$(cd "$(dirname "$0")/../.." && pwd)
probe_app="$RUNNER_TEMP/Geometry.app"
mkdir -p "$probe_app"
cp "$TASK_GEOMETRY_BINARY" "$probe_app/Geometry"
cp "$root/.github/ios-geometry/Info.plist" "$probe_app/Info.plist"
codesign --force --sign - "$probe_app"
xcrun simctl list devices available --json > "$RUNNER_TEMP/geometry-devices.json"
task_simulator_sdk=$(xcrun --sdk iphonesimulator --show-sdk-version)
python3 - "$RUNNER_TEMP/geometry-devices.json" "$task_simulator_sdk" > "$RUNNER_TEMP/geometry-devices.txt" <<'PY'
import json,sys
devices=json.load(open(sys.argv[1]))['devices']
# Use the selected stable Xcode SDK, never an installed beta simulator runtime.
runtimes=['com.apple.CoreSimulator.SimRuntime.iOS-'+sys.argv[2].replace('.','-')]
for kind in ('iPhone','iPad'):
    found=next(((runtime,d) for runtime in runtimes for d in devices.get(runtime,[])
                if d.get('isAvailable') and d['name'].startswith(kind)),None)
    if found is None:
        raise SystemExit('Required UIKit simulator unavailable: '+kind)
    runtime,device=found
    print(kind,device['udid'],device['state'])
    print('UIKit probe: '+runtime+' / '+device['name'],file=sys.stderr)
PY
started=()
cleanup() {
  for simulator in "${started[@]}"; do
    xcrun simctl shutdown "$simulator" >/dev/null 2>&1 || true
  done
}
trap cleanup EXIT
# What the probe said, what the simulator logged about it and any crash report.
diagnose() {
  local simulator=$1 kind=$2 since=$3 report
  echo "::group::Geometry probe output ($kind)"
  cat "$RUNNER_TEMP/geometry-$kind.stdout" "$RUNNER_TEMP/geometry-$kind.stderr" 2>/dev/null || true
  echo "::endgroup::"
  echo "::group::Simulator log of the probe ($kind)"
  xcrun simctl spawn "$simulator" log show --style compact --start "$since" \
    --predicate 'process == "Geometry" OR eventMessage CONTAINS "io.fyne.geometry"' 2>&1 | tail -n 300 || true
  echo "::endgroup::"
  for report in $(ls -t "$HOME"/Library/Logs/DiagnosticReports/Geometry* 2>/dev/null | head -n 2 || true); do
    echo "::group::Crash report $report"
    head -c 20000 "$report" || true
    echo
    echo "::endgroup::"
  done
}
while read -r kind simulator state; do
  printf '%s FYNE_UIKIT_DEVICE kind=%s simulator=%s phase=boot\n' "$(date -u +%FT%TZ)" "$kind" "$simulator"
  if [ "$state" != Booted ]; then
    xcrun simctl boot "$simulator"
  fi
  started+=("$simulator")
  xcrun simctl bootstatus "$simulator" -b
  printf '%s FYNE_UIKIT_DEVICE kind=%s simulator=%s phase=install\n' "$(date -u +%FT%TZ)" "$kind" "$simulator"
  xcrun simctl install "$simulator" "$probe_app"
  # Launched once on its own first: XCTest says only that the label never appeared,
  # and an app that exits at launch (a Go panic leaves no crash report) or traps in
  # UIKit must say why.
  printf '%s FYNE_UIKIT_DEVICE kind=%s simulator=%s phase=launch\n' "$(date -u +%FT%TZ)" "$kind" "$simulator"
  since=$(date '+%Y-%m-%d %H:%M:%S')
  xcrun simctl launch --terminate-running-process \
    --stdout="$RUNNER_TEMP/geometry-$kind.stdout" --stderr="$RUNNER_TEMP/geometry-$kind.stderr" \
    "$simulator" io.fyne.geometry
  sleep 10
  if ! xcrun simctl terminate "$simulator" io.fyne.geometry; then
    echo "The geometry probe exited within 10 s of its launch" >&2
    diagnose "$simulator" "$kind" "$since"
    exit 1
  fi
  printf '%s FYNE_UIKIT_DEVICE kind=%s simulator=%s phase=test\n' "$(date -u +%FT%TZ)" "$kind" "$simulator"
  since=$(date '+%Y-%m-%d %H:%M:%S')
  if ! xcodebuild test -project "$root/.github/ios-geometry/UIKitGeometry.xcodeproj" \
    -scheme GeometryUITests -destination "id=$simulator" -parallel-testing-enabled NO \
    -derivedDataPath "$RUNNER_TEMP/geometry-ui-$kind" \
    -resultBundlePath "$RUNNER_TEMP/geometry-ui-$kind.xcresult" CODE_SIGNING_ALLOWED=NO; then
    diagnose "$simulator" "$kind" "$since"
    exit 1
  fi
  printf '%s FYNE_UIKIT_DEVICE kind=%s simulator=%s phase=passed\n' "$(date -u +%FT%TZ)" "$kind" "$simulator"
  xcrun simctl shutdown "$simulator"
done < "$RUNNER_TEMP/geometry-devices.txt"
