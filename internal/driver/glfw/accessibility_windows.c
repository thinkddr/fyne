#define CINTERFACE
#define COBJMACROS
#include "accessibility_windows.h"
#include <windows.h>
#include <ole2.h>
#include <oleacc.h>
#include <stdio.h>

// ============================================================
// UIA type definitions (manual, for MinGW/CGo compatibility)
// ============================================================

typedef struct FyneUIAElement FyneUIAElement;

// Forward-declared interface structs
typedef struct IRawSimple { struct IRawSimpleVtbl* lpVtbl; } IRawSimple;
typedef struct IRawFragment { struct IRawFragmentVtbl* lpVtbl; } IRawFragment;
typedef struct IRawFragRoot { struct IRawFragRootVtbl* lpVtbl; } IRawFragRoot;

typedef int PROPERTYID;
typedef int PATTERNID;
typedef int EVENTID;

enum UIAProviderOptions {
    UIAProviderOptions_ServerSideProvider = 0x2,
};

enum UIANavigateDirection {
    UIANavigateDirection_Parent = 0,
    UIANavigateDirection_NextSibling = 1,
    UIANavigateDirection_PreviousSibling = 2,
    UIANavigateDirection_FirstChild = 3,
    UIANavigateDirection_LastChild = 4,
};

enum UIAStructureChangeType {
    UIAStructureChangeType_ChildAdded = 0,
    UIAStructureChangeType_ChildRemoved = 1,
    UIAStructureChangeType_ChildrenInvalidated = 2,
    UIAStructureChangeType_ChildrenBulkAdded = 3,
    UIAStructureChangeType_ChildrenBulkRemoved = 4,
    UIAStructureChangeType_ChildrenReordered = 5,
};

typedef struct { double left, top, width, height; } UIARect;

// Vtables
struct IRawSimpleVtbl {
    HRESULT (STDMETHODCALLTYPE *QueryInterface)(IRawSimple*, REFIID, void**);
    ULONG   (STDMETHODCALLTYPE *AddRef)(IRawSimple*);
    ULONG   (STDMETHODCALLTYPE *Release)(IRawSimple*);
    HRESULT (STDMETHODCALLTYPE *get_ProviderOptions)(IRawSimple*, int*);
    HRESULT (STDMETHODCALLTYPE *GetPatternProvider)(IRawSimple*, PATTERNID, IUnknown**);
    HRESULT (STDMETHODCALLTYPE *GetPropertyValue)(IRawSimple*, PROPERTYID, VARIANT*);
    HRESULT (STDMETHODCALLTYPE *get_HostRawElementProvider)(IRawSimple*, IRawSimple**);
};

struct IRawFragmentVtbl {
    HRESULT (STDMETHODCALLTYPE *QueryInterface)(IRawFragment*, REFIID, void**);
    ULONG   (STDMETHODCALLTYPE *AddRef)(IRawFragment*);
    ULONG   (STDMETHODCALLTYPE *Release)(IRawFragment*);
    HRESULT (STDMETHODCALLTYPE *Navigate)(IRawFragment*, int, IRawFragment**);
    HRESULT (STDMETHODCALLTYPE *GetRuntimeId)(IRawFragment*, SAFEARRAY**);
    HRESULT (STDMETHODCALLTYPE *get_BoundingRectangle)(IRawFragment*, UIARect*);
    HRESULT (STDMETHODCALLTYPE *GetEmbeddedFragmentRoots)(IRawFragment*, SAFEARRAY**);
    HRESULT (STDMETHODCALLTYPE *SetFocus)(IRawFragment*);
    HRESULT (STDMETHODCALLTYPE *get_FragmentRoot)(IRawFragment*, IRawFragRoot**);
};

struct IRawFragRootVtbl {
    HRESULT (STDMETHODCALLTYPE *QueryInterface)(IRawFragRoot*, REFIID, void**);
    ULONG   (STDMETHODCALLTYPE *AddRef)(IRawFragRoot*);
    ULONG   (STDMETHODCALLTYPE *Release)(IRawFragRoot*);
    HRESULT (STDMETHODCALLTYPE *ElementProviderFromPoint)(IRawFragRoot*, double, double, IRawSimple**);
    HRESULT (STDMETHODCALLTYPE *GetFocus)(IRawFragRoot*, IRawFragment**);
};

// ============================================================
// Element struct (used for both root and children)
// ============================================================

struct FyneUIAElement {
    IRawSimple    simple;
    IRawFragment  fragment;
    IRawFragRoot  fragRoot;

    LONG refCount;
    int  isRoot;
    HWND hwnd;
    int  uniqueId;

    // Child fields
    FyneUIAElement* parent;
    WCHAR* name;
    int    controlType;
    double x, y, width, height;
    int    childIndex;

    // Root fields
    FyneUIAElement** children;
    int childCount;
    int childCapacity;
};

// Pointer adjustment macros
#define ELEM_FROM_SIMPLE(p)   ((FyneUIAElement*)(p))
#define ELEM_FROM_FRAGMENT(p) ((FyneUIAElement*)((char*)(p) - offsetof(FyneUIAElement, fragment)))
#define ELEM_FROM_FRAGROOT(p) ((FyneUIAElement*)((char*)(p) - offsetof(FyneUIAElement, fragRoot)))

// ============================================================
// GUIDs
// ============================================================

static const IID LOCAL_IID_IUnknown =
    {0x00000000,0x0000,0x0000,{0xC0,0x00,0x00,0x00,0x00,0x00,0x00,0x46}};
static const IID IID_IRawSimple =
    {0xd6dd68d1,0x86fd,0x4332,{0x86,0x66,0x9a,0xbe,0xde,0xa2,0xd2,0x4c}};
static const IID IID_IRawFragment =
    {0xf7063da8,0x8359,0x439c,{0x92,0x97,0xbb,0xc5,0x29,0x9a,0x7d,0x87}};
static const IID IID_IRawFragRoot =
    {0x620ce2a5,0xab8f,0x40a9,{0x86,0xcb,0xde,0x3c,0x75,0x59,0x9b,0x58}};

// UIA property/control IDs
#define UIA_RuntimeIdPropertyId            30000
#define UIA_BoundingRectanglePropertyId    30001
#define UIA_ControlTypePropertyId          30003
#define UIA_LocalizedControlTypePropertyId 30004
#define UIA_NamePropertyId                 30005
#define UIA_HasKeyboardFocusPropertyId     30008
#define UIA_IsKeyboardFocusablePropertyId  30009
#define UIA_IsEnabledPropertyId            30010
#define UIA_AutomationIdPropertyId         30011
#define UIA_IsControlElementPropertyId     30016
#define UIA_IsContentElementPropertyId     30017
#define UIA_NativeWindowHandlePropertyId   30020
#define UIA_ProviderDescriptionPropertyId  30107

#define UIA_WindowControlTypeId      50032
#define UIA_ButtonControlTypeId      50000
#define UIA_HyperlinkControlTypeId   50005
#define UIA_TextControlTypeId        50020
#define UIA_GroupControlTypeId        50026
#define UIA_PaneControlTypeId        50033
#define UIA_CustomControlTypeId      50025

#define UiaAppendRuntimeId  3
#define UiaRootObjectId    (-25)

// UIA Event IDs
#define UIA_AutomationFocusChangedEventId      20005
#define UIA_StructureChangedEventId            20002
#define UIA_Window_WindowOpenedEventId         20016

#define WM_FYNE_RAISE_FOCUS (WM_APP + 100)
#define WM_FYNE_FOCUS_CHILD (WM_APP + 101)

// ============================================================
// Dynamically loaded UIA functions
// ============================================================

typedef LRESULT (WINAPI *PFN_UiaReturnRawElementProvider)(HWND, WPARAM, LPARAM, void*);
typedef HRESULT (WINAPI *PFN_UiaHostProviderFromHwnd)(HWND, void**);
typedef HRESULT (WINAPI *PFN_UiaRaiseAutomationEvent)(void*, EVENTID);
typedef HRESULT (WINAPI *PFN_UiaRaiseStructureChangedEvent)(void*, int, int*, int);
typedef HRESULT (WINAPI *PFN_UiaDisconnectProvider)(void*);

static PFN_UiaReturnRawElementProvider pfnUiaReturn = NULL;
static PFN_UiaHostProviderFromHwnd     pfnUiaHost   = NULL;
static PFN_UiaRaiseAutomationEvent     pfnUiaRaiseEvent = NULL;
static PFN_UiaRaiseStructureChangedEvent pfnUiaRaiseStructure = NULL;
static PFN_UiaDisconnectProvider       pfnUiaDisconnect = NULL;
static HMODULE hUiaCore = NULL;

static void loadUiaFunctions(void) {
    if (hUiaCore) return;
    hUiaCore = LoadLibraryW(L"uiautomationcore.dll");
    fprintf(stderr, "[a11y] LoadLibrary uiautomationcore.dll => %p\n", (void*)hUiaCore);
    if (!hUiaCore) return;
    pfnUiaReturn = (PFN_UiaReturnRawElementProvider)GetProcAddress(hUiaCore, "UiaReturnRawElementProvider");
    pfnUiaHost   = (PFN_UiaHostProviderFromHwnd)GetProcAddress(hUiaCore, "UiaHostProviderFromHwnd");
    pfnUiaRaiseEvent = (PFN_UiaRaiseAutomationEvent)GetProcAddress(hUiaCore, "UiaRaiseAutomationEvent");
    pfnUiaRaiseStructure = (PFN_UiaRaiseStructureChangedEvent)GetProcAddress(hUiaCore, "UiaRaiseStructureChangedEvent");
    pfnUiaDisconnect = (PFN_UiaDisconnectProvider)GetProcAddress(hUiaCore, "UiaDisconnectProvider");
    fprintf(stderr, "[a11y] UiaReturnRawElementProvider => %p\n", (void*)pfnUiaReturn);
    fprintf(stderr, "[a11y] UiaHostProviderFromHwnd => %p\n", (void*)pfnUiaHost);
    fprintf(stderr, "[a11y] UiaRaiseAutomationEvent => %p\n", (void*)pfnUiaRaiseEvent);
    fprintf(stderr, "[a11y] UiaRaiseStructureChangedEvent => %p\n", (void*)pfnUiaRaiseStructure);
    fprintf(stderr, "[a11y] UiaDisconnectProvider => %p\n", (void*)pfnUiaDisconnect);
}

// ============================================================
// Globals
// ============================================================

static FyneUIAElement* g_root = NULL;
static HWND g_hwnd = NULL;
static WNDPROC g_origWndProc = NULL;
static struct IRawSimpleVtbl   g_simpleVtbl;
static struct IRawFragmentVtbl g_fragmentVtbl;
static struct IRawFragRootVtbl g_fragRootVtbl;
static int g_vtblInit = 0;
static int g_nextId = 1;
static int g_focusedIndex = -1; // index of focused child, -1 = none

// ============================================================
// Helpers
// ============================================================

static WCHAR* utf8ToWide(const char* utf8) {
    if (!utf8 || !utf8[0]) {
        WCHAR* e = (WCHAR*)malloc(sizeof(WCHAR));
        if (e) e[0] = L'\0';
        return e;
    }
    int len = MultiByteToWideChar(CP_UTF8, 0, utf8, -1, NULL, 0);
    WCHAR* w = (WCHAR*)malloc(len * sizeof(WCHAR));
    if (w) MultiByteToWideChar(CP_UTF8, 0, utf8, -1, w, len);
    return w;
}

static int roleToUIA(WinAccessibilityRole role) {
    switch (role) {
    case WinAccessibilityRoleButton: return UIA_ButtonControlTypeId;
    case WinAccessibilityRoleText:   return UIA_TextControlTypeId;
    case WinAccessibilityRoleLink:   return UIA_HyperlinkControlTypeId;
    case WinAccessibilityRoleGroup:  return UIA_GroupControlTypeId;
    default:                         return UIA_PaneControlTypeId;
    }
}

// Core QueryInterface shared by all interfaces
static HRESULT elemQI(FyneUIAElement* elem, REFIID riid, void** ppv) {
    if (!ppv) return E_POINTER;
    if (IsEqualIID(riid, &LOCAL_IID_IUnknown) || IsEqualIID(riid, &IID_IRawSimple)) {
        *ppv = &elem->simple;
        InterlockedIncrement(&elem->refCount);
        return S_OK;
    }
    if (IsEqualIID(riid, &IID_IRawFragment)) {
        *ppv = &elem->fragment;
        InterlockedIncrement(&elem->refCount);
        return S_OK;
    }
    if (elem->isRoot && IsEqualIID(riid, &IID_IRawFragRoot)) {
        *ppv = &elem->fragRoot;
        InterlockedIncrement(&elem->refCount);
        return S_OK;
    }
    fprintf(stderr, "[a11y] QI unknown IID {%08lx-%04x-%04x-...} isRoot=%d\n",
        riid->Data1, riid->Data2, riid->Data3, elem->isRoot);
    *ppv = NULL;
    return E_NOINTERFACE;
}

// ============================================================
// IRawElementProviderSimple implementation
// ============================================================

static HRESULT STDMETHODCALLTYPE S_QI(IRawSimple* This, REFIID riid, void** ppv) {
    return elemQI(ELEM_FROM_SIMPLE(This), riid, ppv);
}
static ULONG STDMETHODCALLTYPE S_AddRef(IRawSimple* This) {
    return InterlockedIncrement(&ELEM_FROM_SIMPLE(This)->refCount);
}
static ULONG STDMETHODCALLTYPE S_Release(IRawSimple* This) {
    FyneUIAElement* e = ELEM_FROM_SIMPLE(This);
    ULONG c = InterlockedDecrement(&e->refCount);
    if (c == 0 && !e->isRoot) { free(e->name); free(e); }
    return c;
}

static HRESULT STDMETHODCALLTYPE S_get_ProviderOptions(IRawSimple* This, int* pRetVal) {
    if (!pRetVal) return E_POINTER;
    *pRetVal = UIAProviderOptions_ServerSideProvider | 0x20 /*UseComThreading*/;
    return S_OK;
}

static HRESULT STDMETHODCALLTYPE S_GetPatternProvider(IRawSimple* This, PATTERNID id, IUnknown** pRetVal) {
    if (!pRetVal) return E_POINTER;
    *pRetVal = NULL;
    return S_OK;
}

static HRESULT STDMETHODCALLTYPE S_GetPropertyValue(IRawSimple* This, PROPERTYID pid, VARIANT* pRetVal) {
    if (!pRetVal) return E_POINTER;
    VariantInit(pRetVal);
    FyneUIAElement* e = ELEM_FROM_SIMPLE(This);

    // For root element, let host provider handle most properties
    // but override focus-related ones (host returns IsKeyboardFocusable=False)
    if (e->isRoot) {
        switch (pid) {
        case UIA_IsKeyboardFocusablePropertyId:
            pRetVal->vt = VT_BOOL;
            pRetVal->boolVal = VARIANT_TRUE;
            fprintf(stderr, "[a11y] Root: IsKeyboardFocusable => TRUE\n");
            break;
        case UIA_HasKeyboardFocusPropertyId:
            pRetVal->vt = VT_BOOL;
            pRetVal->boolVal = (GetForegroundWindow() == e->hwnd) ? VARIANT_TRUE : VARIANT_FALSE;
            fprintf(stderr, "[a11y] Root: HasKeyboardFocus => %s\n",
                pRetVal->boolVal == VARIANT_TRUE ? "TRUE" : "FALSE");
            break;
        case UIA_IsControlElementPropertyId:
        case UIA_IsContentElementPropertyId:
        case UIA_IsEnabledPropertyId:
            pRetVal->vt = VT_BOOL;
            pRetVal->boolVal = VARIANT_TRUE;
            break;
        case UIA_ProviderDescriptionPropertyId:
            pRetVal->vt = VT_BSTR;
            pRetVal->bstrVal = SysAllocString(L"Fyne Accessibility Provider");
            break;
        default:
            // Return VT_EMPTY - UIA will fall back to host provider
            break;
        }
        return S_OK;
    }

    // Child element properties
    switch (pid) {
    case UIA_ControlTypePropertyId:
        pRetVal->vt = VT_I4;
        pRetVal->lVal = e->controlType;
        fprintf(stderr, "[a11y] GetPropertyValue ControlType=%ld child[%d]\n", pRetVal->lVal, e->childIndex);
        break;
    case UIA_NamePropertyId:
        pRetVal->vt = VT_BSTR;
        pRetVal->bstrVal = SysAllocString(e->name);
        fprintf(stderr, "[a11y] GetPropertyValue Name child[%d]\n", e->childIndex);
        break;
    case UIA_AutomationIdPropertyId: {
        WCHAR buf[32];
        wsprintfW(buf, L"fyne_%d", e->uniqueId);
        pRetVal->vt = VT_BSTR;
        pRetVal->bstrVal = SysAllocString(buf);
        break;
    }
    case UIA_IsControlElementPropertyId:
    case UIA_IsContentElementPropertyId:
    case UIA_IsEnabledPropertyId:
        pRetVal->vt = VT_BOOL;
        pRetVal->boolVal = VARIANT_TRUE;
        break;
    case UIA_IsKeyboardFocusablePropertyId:
        pRetVal->vt = VT_BOOL;
        pRetVal->boolVal = VARIANT_TRUE;
        break;
    case UIA_HasKeyboardFocusPropertyId:
        pRetVal->vt = VT_BOOL;
        pRetVal->boolVal = (g_focusedIndex == e->childIndex) ? VARIANT_TRUE : VARIANT_FALSE;
        break;
    case UIA_ProviderDescriptionPropertyId:
        pRetVal->vt = VT_BSTR;
        pRetVal->bstrVal = SysAllocString(L"Fyne Accessibility Provider");
        break;
    }
    return S_OK;
}

static HRESULT STDMETHODCALLTYPE S_get_HostRawElementProvider(IRawSimple* This, IRawSimple** pRetVal) {
    if (!pRetVal) return E_POINTER;
    *pRetVal = NULL;
    FyneUIAElement* e = ELEM_FROM_SIMPLE(This);
    if (e->isRoot && pfnUiaHost) {
        HRESULT hr = pfnUiaHost(e->hwnd, (void**)pRetVal);
        fprintf(stderr, "[a11y] get_HostRawElementProvider hr=0x%lx result=%p\n", hr, (void*)*pRetVal);
    } else {
        fprintf(stderr, "[a11y] get_HostRawElementProvider (child) => NULL\n");
    }
    return S_OK;
}

// ============================================================
// IRawElementProviderFragment implementation
// ============================================================

static HRESULT STDMETHODCALLTYPE F_QI(IRawFragment* This, REFIID riid, void** ppv) {
    return elemQI(ELEM_FROM_FRAGMENT(This), riid, ppv);
}
static ULONG STDMETHODCALLTYPE F_AddRef(IRawFragment* This) {
    return InterlockedIncrement(&ELEM_FROM_FRAGMENT(This)->refCount);
}
static ULONG STDMETHODCALLTYPE F_Release(IRawFragment* This) {
    FyneUIAElement* e = ELEM_FROM_FRAGMENT(This);
    ULONG c = InterlockedDecrement(&e->refCount);
    if (c == 0 && !e->isRoot) { free(e->name); free(e); }
    return c;
}

static HRESULT STDMETHODCALLTYPE F_Navigate(IRawFragment* This, int direction, IRawFragment** pRetVal) {
    if (!pRetVal) return E_POINTER;
    *pRetVal = NULL;
    FyneUIAElement* e = ELEM_FROM_FRAGMENT(This);
    FyneUIAElement* target = NULL;

    if (e->isRoot) {
        switch (direction) {
        case UIANavigateDirection_FirstChild:
            if (e->childCount > 0) target = e->children[0];
            fprintf(stderr, "[a11y] Navigate FirstChild from root => %p (children=%d)\n",
                (void*)target, e->childCount);
            break;
        case UIANavigateDirection_LastChild:
            if (e->childCount > 0) target = e->children[e->childCount - 1];
            fprintf(stderr, "[a11y] Navigate LastChild from root => %p\n", (void*)target);
            break;
        case UIANavigateDirection_Parent:
            fprintf(stderr, "[a11y] Navigate Parent from root => NULL\n");
            break;
        default:
            fprintf(stderr, "[a11y] Navigate dir=%d from root => NULL\n", direction);
            break;
        }
    } else {
        FyneUIAElement* p = e->parent;
        switch (direction) {
        case UIANavigateDirection_Parent:
            target = p;
            fprintf(stderr, "[a11y] Navigate Parent from child[%d] => root\n", e->childIndex);
            break;
        case UIANavigateDirection_NextSibling:
            if (p && e->childIndex + 1 < p->childCount)
                target = p->children[e->childIndex + 1];
            fprintf(stderr, "[a11y] Navigate NextSibling from child[%d] => %p\n",
                e->childIndex, (void*)target);
            break;
        case UIANavigateDirection_PreviousSibling:
            if (p && e->childIndex > 0)
                target = p->children[e->childIndex - 1];
            fprintf(stderr, "[a11y] Navigate PrevSibling from child[%d] => %p\n",
                e->childIndex, (void*)target);
            break;
        case UIANavigateDirection_FirstChild:
        case UIANavigateDirection_LastChild:
            // Leaf elements have no children
            fprintf(stderr, "[a11y] Navigate child from leaf[%d] => NULL\n", e->childIndex);
            break;
        }
    }

    if (target) {
        *pRetVal = &target->fragment;
        InterlockedIncrement(&target->refCount);
    }
    return S_OK;
}

static HRESULT STDMETHODCALLTYPE F_GetRuntimeId(IRawFragment* This, SAFEARRAY** pRetVal) {
    if (!pRetVal) return E_POINTER;
    FyneUIAElement* e = ELEM_FROM_FRAGMENT(This);

    SAFEARRAYBOUND bound = {2, 0};
    SAFEARRAY* psa = SafeArrayCreate(VT_I4, 1, &bound);
    if (!psa) return E_OUTOFMEMORY;

    long idx = 0;
    int val = UiaAppendRuntimeId;
    SafeArrayPutElement(psa, &idx, &val);
    idx = 1;
    val = e->uniqueId;
    SafeArrayPutElement(psa, &idx, &val);

    *pRetVal = psa;
    return S_OK;
}

static HRESULT STDMETHODCALLTYPE F_get_BoundingRectangle(IRawFragment* This, UIARect* pRetVal) {
    if (!pRetVal) return E_POINTER;
    FyneUIAElement* e = ELEM_FROM_FRAGMENT(This);

    if (e->isRoot) {
        RECT rc;
        GetClientRect(e->hwnd, &rc);
        POINT pt = {0, 0};
        ClientToScreen(e->hwnd, &pt);
        pRetVal->left   = pt.x;
        pRetVal->top    = pt.y;
        pRetVal->width  = rc.right - rc.left;
        pRetVal->height = rc.bottom - rc.top;
    } else {
        POINT pt = {(LONG)e->x, (LONG)e->y};
        ClientToScreen(e->hwnd, &pt);
        pRetVal->left   = pt.x;
        pRetVal->top    = pt.y;
        pRetVal->width  = e->width;
        pRetVal->height = e->height;
    }
    return S_OK;
}

static HRESULT STDMETHODCALLTYPE F_GetEmbeddedFragmentRoots(IRawFragment* This, SAFEARRAY** pRetVal) {
    if (pRetVal) *pRetVal = NULL;
    return S_OK;
}

static HRESULT STDMETHODCALLTYPE F_SetFocus(IRawFragment* This) {
    FyneUIAElement* e = ELEM_FROM_FRAGMENT(This);
    if (!e->isRoot) {
        g_focusedIndex = e->childIndex;
        fprintf(stderr, "[a11y] SetFocus on child[%d]\n", e->childIndex);
        if (pfnUiaRaiseEvent) {
            pfnUiaRaiseEvent(&e->simple, UIA_AutomationFocusChangedEventId);
        }
    }
    return S_OK;
}

static HRESULT STDMETHODCALLTYPE F_get_FragmentRoot(IRawFragment* This, IRawFragRoot** pRetVal) {
    if (!pRetVal) return E_POINTER;
    *pRetVal = NULL;
    FyneUIAElement* e = ELEM_FROM_FRAGMENT(This);
    FyneUIAElement* root = e->isRoot ? e : e->parent;
    if (root) {
        *pRetVal = &root->fragRoot;
        InterlockedIncrement(&root->refCount);
    }
    return S_OK;
}

// ============================================================
// IRawElementProviderFragmentRoot implementation
// ============================================================

static HRESULT STDMETHODCALLTYPE FR_QI(IRawFragRoot* This, REFIID riid, void** ppv) {
    return elemQI(ELEM_FROM_FRAGROOT(This), riid, ppv);
}
static ULONG STDMETHODCALLTYPE FR_AddRef(IRawFragRoot* This) {
    return InterlockedIncrement(&ELEM_FROM_FRAGROOT(This)->refCount);
}
static ULONG STDMETHODCALLTYPE FR_Release(IRawFragRoot* This) {
    return InterlockedDecrement(&ELEM_FROM_FRAGROOT(This)->refCount);
}

static HRESULT STDMETHODCALLTYPE FR_ElementProviderFromPoint(IRawFragRoot* This,
    double x, double y, IRawSimple** pRetVal) {
    if (!pRetVal) return E_POINTER;
    *pRetVal = NULL;
    FyneUIAElement* e = ELEM_FROM_FRAGROOT(This);

    POINT pt = {(LONG)x, (LONG)y};
    ScreenToClient(e->hwnd, &pt);

    for (int i = 0; i < e->childCount; i++) {
        FyneUIAElement* c = e->children[i];
        if (pt.x >= c->x && pt.x < c->x + c->width &&
            pt.y >= c->y && pt.y < c->y + c->height) {
            *pRetVal = &c->simple;
            InterlockedIncrement(&c->refCount);
            fprintf(stderr, "[a11y] ElementProviderFromPoint (%g,%g) => child[%d]\n", x, y, i);
            return S_OK;
        }
    }
    fprintf(stderr, "[a11y] ElementProviderFromPoint (%g,%g) => none\n", x, y);
    return S_OK;
}

static HRESULT STDMETHODCALLTYPE FR_GetFocus(IRawFragRoot* This, IRawFragment** pRetVal) {
    if (!pRetVal) return E_POINTER;
    *pRetVal = NULL;
    FyneUIAElement* e = ELEM_FROM_FRAGROOT(This);
    if (g_focusedIndex >= 0 && g_focusedIndex < e->childCount) {
        FyneUIAElement* child = e->children[g_focusedIndex];
        *pRetVal = &child->fragment;
        InterlockedIncrement(&child->refCount);
        fprintf(stderr, "[a11y] GetFocus => child[%d]\n", g_focusedIndex);
    }
    return S_OK;
}

// ============================================================
// Vtable setup
// ============================================================

static void initVtbls(void) {
    if (g_vtblInit) return;

    g_simpleVtbl.QueryInterface = S_QI;
    g_simpleVtbl.AddRef = S_AddRef;
    g_simpleVtbl.Release = S_Release;
    g_simpleVtbl.get_ProviderOptions = S_get_ProviderOptions;
    g_simpleVtbl.GetPatternProvider = S_GetPatternProvider;
    g_simpleVtbl.GetPropertyValue = S_GetPropertyValue;
    g_simpleVtbl.get_HostRawElementProvider = S_get_HostRawElementProvider;

    g_fragmentVtbl.QueryInterface = F_QI;
    g_fragmentVtbl.AddRef = F_AddRef;
    g_fragmentVtbl.Release = F_Release;
    g_fragmentVtbl.Navigate = F_Navigate;
    g_fragmentVtbl.GetRuntimeId = F_GetRuntimeId;
    g_fragmentVtbl.get_BoundingRectangle = F_get_BoundingRectangle;
    g_fragmentVtbl.GetEmbeddedFragmentRoots = F_GetEmbeddedFragmentRoots;
    g_fragmentVtbl.SetFocus = F_SetFocus;
    g_fragmentVtbl.get_FragmentRoot = F_get_FragmentRoot;

    g_fragRootVtbl.QueryInterface = FR_QI;
    g_fragRootVtbl.AddRef = FR_AddRef;
    g_fragRootVtbl.Release = FR_Release;
    g_fragRootVtbl.ElementProviderFromPoint = FR_ElementProviderFromPoint;
    g_fragRootVtbl.GetFocus = FR_GetFocus;

    g_vtblInit = 1;
}

static FyneUIAElement* createElement(int isRoot, HWND hwnd) {
    FyneUIAElement* e = (FyneUIAElement*)calloc(1, sizeof(FyneUIAElement));
    if (!e) return NULL;
    e->simple.lpVtbl   = &g_simpleVtbl;
    e->fragment.lpVtbl  = &g_fragmentVtbl;
    e->fragRoot.lpVtbl  = &g_fragRootVtbl;
    e->refCount = 1;
    e->isRoot = isRoot;
    e->hwnd = hwnd;
    e->uniqueId = g_nextId++;
    return e;
}

// ============================================================
// Window subclass
// ============================================================

static void focusChild(int index) {
    if (!g_root || index < 0 || index >= g_root->childCount) return;
    if (index == g_focusedIndex) return;
    // Defer the focus event via PostMessage to avoid re-entrancy delays
    PostMessageW(g_hwnd, WM_FYNE_FOCUS_CHILD, (WPARAM)index, 0);
}

static int hitTestChild(int clientX, int clientY) {
    if (!g_root) return -1;
    for (int i = 0; i < g_root->childCount; i++) {
        FyneUIAElement* c = g_root->children[i];
        if (clientX >= c->x && clientX < c->x + c->width &&
            clientY >= c->y && clientY < c->y + c->height) {
            return i;
        }
    }
    return -1;
}

static LRESULT CALLBACK AccessibilityWndProc(HWND hwnd, UINT msg, WPARAM wParam, LPARAM lParam) {
    if (msg == WM_GETOBJECT) {
        fprintf(stderr, "[a11y] WM_GETOBJECT lParam=%lld wParam=0x%llx\n",
            (long long)lParam, (unsigned long long)wParam);
        if (lParam == (LPARAM)UiaRootObjectId) {
            if (g_root && pfnUiaReturn) {
                fprintf(stderr, "[a11y] Calling UiaReturnRawElementProvider root=%p children=%d\n",
                    (void*)g_root, g_root->childCount);
                LRESULT lr = pfnUiaReturn(hwnd, wParam, lParam, &g_root->simple);
                fprintf(stderr, "[a11y] UiaReturnRawElementProvider returned %lld\n", (long long)lr);
                return lr;
            }
        }
        // For OBJID_CLIENT and other object IDs, pass to DefWindowProc
        // to allow UIA to also use the host provider
        return DefWindowProcW(hwnd, msg, wParam, lParam);
    }

    if (msg == WM_SETFOCUS || msg == WM_ACTIVATE) {
        fprintf(stderr, "[a11y] %s wParam=0x%llx\n",
            msg == WM_SETFOCUS ? "WM_SETFOCUS" : "WM_ACTIVATE",
            (unsigned long long)wParam);
        // Raise UIA focus event when window gains focus
        if (msg == WM_SETFOCUS || (msg == WM_ACTIVATE && LOWORD(wParam) != 0 /*WA_INACTIVE*/)) {
            // Defer so UIA has time to integrate provider
            PostMessageW(hwnd, WM_FYNE_RAISE_FOCUS, 0, 0);
        }
    }

    if (msg == WM_FYNE_RAISE_FOCUS) {
        fprintf(stderr, "[a11y] WM_FYNE_RAISE_FOCUS: raising focus event\n");
        if (g_root && pfnUiaRaiseEvent) {
            pfnUiaRaiseEvent(&g_root->simple, UIA_AutomationFocusChangedEventId);
        }
        return 0;
    }

    if (msg == WM_FYNE_FOCUS_CHILD) {
        int index = (int)wParam;
        if (g_root && index >= 0 && index < g_root->childCount) {
            g_focusedIndex = index;
            FyneUIAElement* child = g_root->children[index];
            fprintf(stderr, "[a11y] FocusChild => child[%d] '%ls'\n", index, child->name);
            if (pfnUiaRaiseEvent) {
                pfnUiaRaiseEvent(&child->simple, UIA_AutomationFocusChangedEventId);
            }
        }
        return 0;
    }

    // Handle mouse clicks - focus the child under cursor
    if (msg == WM_LBUTTONDOWN) {
        int x = (int)(short)LOWORD(lParam);
        int y = (int)(short)HIWORD(lParam);
        int hit = hitTestChild(x, y);
        if (hit >= 0) {
            focusChild(hit);
        }
    }

    // Handle Tab key - cycle focus between children
    if (msg == WM_KEYDOWN && wParam == VK_TAB) {
        if (g_root && g_root->childCount > 0) {
            int next;
            if (GetKeyState(VK_SHIFT) & 0x8000) {
                // Shift+Tab: previous
                next = (g_focusedIndex <= 0) ? g_root->childCount - 1 : g_focusedIndex - 1;
            } else {
                // Tab: next
                next = (g_focusedIndex + 1) % g_root->childCount;
            }
            focusChild(next);
        }
    }

    return CallWindowProcW(g_origWndProc, hwnd, msg, wParam, lParam);
}

// ============================================================
// Public API
// ============================================================

void WinAccessibilitySetWindow(void* hwnd) {
    HWND h = (HWND)hwnd;
    if (h == g_hwnd && g_root) return;

    loadUiaFunctions();

    if (g_hwnd && g_origWndProc) {
        SetWindowLongPtrW(g_hwnd, GWLP_WNDPROC, (LONG_PTR)g_origWndProc);
        g_origWndProc = NULL;
    }

    g_hwnd = h;
    initVtbls();

    if (!g_root) {
        g_root = createElement(1, h);
        if (!g_root) return;
    }
    g_root->hwnd = h;

    g_origWndProc = (WNDPROC)SetWindowLongPtrW(h, GWLP_WNDPROC, (LONG_PTR)AccessibilityWndProc);
    fprintf(stderr, "[a11y] Window subclassed: hwnd=%p origProc=%p\n", (void*)h, (void*)g_origWndProc);
}

void WinAccessibilityAddElement(const char* name, WinAccessibilityRole role,
    double x, double y, double width, double height) {
    if (!g_root) return;

    if (g_root->childCount >= g_root->childCapacity) {
        int newCap = g_root->childCapacity == 0 ? 16 : g_root->childCapacity * 2;
        FyneUIAElement** a = (FyneUIAElement**)realloc(g_root->children, newCap * sizeof(FyneUIAElement*));
        if (!a) return;
        g_root->children = a;
        g_root->childCapacity = newCap;
    }

    FyneUIAElement* child = createElement(0, g_root->hwnd);
    if (!child) return;
    child->parent = g_root;
    child->name = utf8ToWide(name);
    child->controlType = roleToUIA(role);
    child->x = x;
    child->y = y;
    child->width = width;
    child->height = height;
    child->childIndex = g_root->childCount;

    g_root->children[g_root->childCount] = child;
    g_root->childCount++;
    fprintf(stderr, "[a11y] AddElement '%s' role=%d at (%.0f,%.0f,%.0f,%.0f) => child[%d]\n",
        name, role, x, y, width, height, child->childIndex);
}

void WinAccessibilityClearElements(void) {
    if (!g_root) return;
    for (int i = 0; i < g_root->childCount; i++) {
        FyneUIAElement* c = g_root->children[i];
        c->parent = NULL;
        c->simple.lpVtbl->Release(&c->simple);
    }
    g_root->childCount = 0;
    g_focusedIndex = -1;
}

void WinAccessibilityUpdate(void) {
    if (!g_root || !g_hwnd) return;
    fprintf(stderr, "[a11y] Update: %d children\n", g_root->childCount);

    // Raise UIA structure changed event
    if (pfnUiaRaiseStructure) {
        int runtimeId[2] = { UiaAppendRuntimeId, g_root->uniqueId };
        HRESULT hr = pfnUiaRaiseStructure(&g_root->simple,
            UIAStructureChangeType_ChildrenInvalidated, runtimeId, 2);
        fprintf(stderr, "[a11y] UiaRaiseStructureChangedEvent => 0x%lx\n", hr);
    }

    // Defer focus event so UIA has time to process the tree
    PostMessageW(g_hwnd, WM_FYNE_RAISE_FOCUS, 0, 0);
}

void WinAccessibilityCleanup(void) {
    if (g_hwnd && g_origWndProc) {
        SetWindowLongPtrW(g_hwnd, GWLP_WNDPROC, (LONG_PTR)g_origWndProc);
        g_origWndProc = NULL;
    }
    if (g_root) {
        if (pfnUiaDisconnect) {
            pfnUiaDisconnect(&g_root->simple);
        }
        WinAccessibilityClearElements();
        free(g_root->children);
        free(g_root);
        g_root = NULL;
    }
    g_hwnd = NULL;
}
