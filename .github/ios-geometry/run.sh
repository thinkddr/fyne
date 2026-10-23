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
python3 - "$RUNNER_TEMP/geometry-devices.json" > "$RUNNER_TEMP/geometry-devices.txt" <<'PY'
import json,re,sys
devices=json.load(open(sys.argv[1]))['devices']
runtimes=sorted((r for r in devices if '.iOS-' in r),key=lambda r:tuple(map(int,re.findall(r'\d+',r))),reverse=True)
for kind in ('iPhone','iPad'):
    found=next(((runtime,d) for runtime in runtimes for d in devices[runtime]
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
while read -r kind simulator state; do
  if [ "$state" != Booted ]; then
    xcrun simctl boot "$simulator"
  fi
  started+=("$simulator")
  xcrun simctl bootstatus "$simulator" -b
  xcrun simctl install "$simulator" "$probe_app"
  xcodebuild test -project "$root/.github/ios-geometry/UIKitGeometry.xcodeproj" \
    -scheme GeometryUITests -destination "id=$simulator" -parallel-testing-enabled NO \
    -derivedDataPath "$RUNNER_TEMP/geometry-ui-$kind" \
    -resultBundlePath "$RUNNER_TEMP/geometry-ui-$kind.xcresult" CODE_SIGNING_ALLOWED=NO
  xcrun simctl shutdown "$simulator"
done < "$RUNNER_TEMP/geometry-devices.txt"
