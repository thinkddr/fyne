#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2026 Daniel Delgado Ramos
"""Exercise the production Android result handler and JNI callback on a JVM.

Compile the activity against a real Android SDK. The runtime fixtures only model
Activity, provider grants and result data; they contain no picker decisions.
"""
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

APP = Path(__file__).resolve().parents[2]

FIXTURES = {
    "android/content/Context.java": """package android.content;
public class Context {
    private final ContentResolver resolver = new ContentResolver();
    public ContentResolver getContentResolver() { return resolver; }
}
""",
    "android/app/Activity.java": """package android.app;
public class Activity extends android.content.Context {
    public static final int RESULT_OK = -1;
}
""",
    "android/app/NativeActivity.java": """package android.app;
public class NativeActivity extends Activity {}
""",
    "android/content/ContentResolver.java": """package android.content;
public class ContentResolver {
    public final java.util.List<String> attempts = new java.util.ArrayList<>();
    public final java.util.List<String> grants = new java.util.ArrayList<>();
    public String refuse;
    public void takePersistableUriPermission(android.net.Uri uri, int flags) {
        attempts.add(uri.toString());
        if (uri.toString().equals(refuse)) { throw new SecurityException("provider refused"); }
        grants.add(uri.toString() + ":" + flags);
    }
}
""",
    "android/net/Uri.java": """package android.net;
public class Uri {
    private final String value;
    private Uri(String value) { this.value = value; }
    public static Uri parse(String value) { return new Uri(value); }
    public String toString() { return value; }
}
""",
    "android/content/Intent.java": """package android.content;
public class Intent {
    public static final int FLAG_GRANT_READ_URI_PERMISSION = 1;
    public static final int FLAG_GRANT_WRITE_URI_PERMISSION = 2;
    public static final int FLAG_GRANT_PERSISTABLE_URI_PERMISSION = 64;
    private int flags;
    private android.net.Uri uri;
    private ClipData clip;
    public Intent setFlags(int flags) { this.flags = flags; return this; }
    public int getFlags() { return flags; }
    public Intent setData(android.net.Uri uri) { this.uri = uri; return this; }
    public android.net.Uri getData() { return uri; }
    public void setClipData(ClipData clip) { this.clip = clip; }
    public ClipData getClipData() { return clip; }
}
""",
    "android/content/ClipData.java": """package android.content;
public class ClipData {
    private final java.util.List<Item> items = new java.util.ArrayList<>();
    public static ClipData newRawUri(CharSequence label, android.net.Uri uri) {
        ClipData data = new ClipData(); data.addItem(new Item(uri)); return data;
    }
    public void addItem(Item item) { items.add(item); }
    public int getItemCount() { return items.size(); }
    public Item getItemAt(int index) { return items.get(index); }
    public static class Item {
        private final android.net.Uri uri;
        public Item(android.net.Uri uri) { this.uri = uri; }
        public android.net.Uri getUri() { return uri; }
    }
}
""",
    "android/util/Log.java": """package android.util;
public class Log {
    public static int w(String tag, String message) { return 0; }
}
""",
    "org/golang/app/PickerTest.java": """package org.golang.app;
import android.app.Activity;
import android.content.ClipData;
import android.content.ContentResolver;
import android.content.Intent;
import android.net.Uri;
import java.util.Arrays;
import java.util.Collections;

public final class PickerTest {
    private static String returned;
    private static int callbacks, checks;
    public static void record(String value) { returned = value; callbacks++; }
    private static void check(boolean ok, String message) {
        checks++; if (!ok) { throw new AssertionError(message); }
    }
    private static Intent single(int flags) {
        return new Intent().setFlags(flags).setData(Uri.parse("content://provider/one"));
    }
    private static Intent multi(int flags) {
        Intent data = new Intent().setFlags(flags);
        ClipData clip = ClipData.newRawUri("selection", Uri.parse("content://provider/one"));
        clip.addItem(new ClipData.Item(Uri.parse("content://provider/two")));
        data.setClipData(clip); return data;
    }
    private static GoNativeActivity activity() {
        returned = null; callbacks = 0; return new GoNativeActivity();
    }
    public static void main(String[] args) {
        System.load(args[0]);
        int read = Intent.FLAG_GRANT_READ_URI_PERMISSION;
        int persist = Intent.FLAG_GRANT_PERSISTABLE_URI_PERMISSION;
        GoNativeActivity a = activity();
        a.onActivityResult(1, Activity.RESULT_OK, single(read | persist));
        check(callbacks == 1 && "content://provider/one".equals(returned), "single URI callback preserved");
        check(a.getContentResolver().grants.equals(Collections.singletonList("content://provider/one:1")),
                "single URI persists read access only");

        a = activity();
        a.onActivityResult(1, Activity.RESULT_OK, multi(read | persist));
        check(callbacks == 1 && "content://provider/one\\ncontent://provider/two".equals(returned),
                "multi URI order and one JNI callback preserved");
        check(a.getContentResolver().grants.equals(Arrays.asList("content://provider/one:1", "content://provider/two:1")),
                "every selected URI persists read access");

        for (int flags : new int[] {0, read, persist, persist | Intent.FLAG_GRANT_WRITE_URI_PERMISSION}) {
            for (boolean multiple : new boolean[] {false, true}) {
                a = activity();
                a.onActivityResult(1, Activity.RESULT_OK, multiple ? multi(flags) : single(flags));
                check(callbacks == 1 && returned.startsWith("content://provider/one"), "temporary result survives missing grants");
                check(a.getContentResolver().attempts.isEmpty(), "no persistence requested without both result flags");
            }
        }

        for (boolean multiple : new boolean[] {false, true}) {
            a = activity();
            ContentResolver resolver = a.getContentResolver();
            resolver.refuse = "content://provider/one";
            a.onActivityResult(1, Activity.RESULT_OK, multiple ? multi(read | persist) : single(read | persist));
            check(callbacks == 1 && returned.startsWith("content://provider/one"), "provider refusal preserves result callback");
            check(resolver.attempts.contains("content://provider/one"), "provider refusal was exercised");
            check(multiple ? resolver.grants.equals(Collections.singletonList("content://provider/two:1")) : resolver.grants.isEmpty(),
                    "provider refusal cannot prevent the remaining URI grant");
        }

        a = activity();
        a.onActivityResult(2, Activity.RESULT_OK, single(read | persist | Intent.FLAG_GRANT_WRITE_URI_PERMISSION));
        check(callbacks == 1 && a.getContentResolver().grants.equals(Collections.singletonList("content://provider/one:1")),
                "save result does not escalate persistence to write access");
        a = activity();
        a.onActivityResult(1, 0, null);
        check(callbacks == 1 && "".equals(returned) && a.getContentResolver().attempts.isEmpty(), "cancel returns empty without provider access");
        a = activity();
        a.onActivityResult(99, Activity.RESULT_OK, single(read | persist));
        check(callbacks == 0 && a.getContentResolver().attempts.isEmpty(), "unrelated activity result is untouched");
        System.out.println("Android picker: " + checks + " functional checks passed");
    }
}
""",
}

JNI = """#include <jni.h>
JNIEXPORT void JNICALL Java_org_golang_app_GoNativeActivity_filePickerReturned(JNIEnv *env, jobject self, jstring value) {
    (void)self;
    jclass recorder = (*env)->FindClass(env, "org/golang/app/PickerTest");
    jmethodID record = (*env)->GetStaticMethodID(env, recorder, "record", "(Ljava/lang/String;)V");
    (*env)->CallStaticVoidMethod(env, recorder, record, value);
}
"""


def run(*args):
    subprocess.run([str(arg) for arg in args], check=True)


def main():
    if len(sys.argv) != 2 or not Path(sys.argv[1]).is_file():
        sys.exit("usage: test_picker.py /path/to/android.jar")
    sdk = Path(sys.argv[1]).resolve()
    java_home = Path(shutil.which("javac")).resolve().parents[1]
    with tempfile.TemporaryDirectory(prefix="fyne-picker-") as temporary:
        root = Path(temporary)
        classes, fixtures = root / "classes", root / "fixtures"
        classes.mkdir()
        fixtures.mkdir()
        run("javac", "--release", "17", "-encoding", "UTF-8", "-cp", sdk, "-d", classes,
            APP / "GoNativeActivity.java", APP / "FyneNotificationReceiver.java")
        paths = []
        for name, content in FIXTURES.items():
            path = fixtures / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(content)
            paths.append(path)
        run("javac", "--release", "17", "-encoding", "UTF-8", "-cp", str(classes) + os.pathsep + str(sdk),
            "-d", fixtures, *paths)
        native = root / "picker.c"
        native.write_text(JNI)
        library = root / "libpicker.so"
        run(os.environ.get("CC", "cc"), "-shared", "-fPIC", "-I" + str(java_home / "include"),
            "-I" + str(java_home / "include/linux"), "-o", library, native)
        run("java", "-cp", os.pathsep.join(map(str, (fixtures, classes, sdk))), "org.golang.app.PickerTest", library)


if __name__ == "__main__":
    main()
