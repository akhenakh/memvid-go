use libc::{c_char, size_t};
use memvid_core::{Memvid, PutOptions, SearchRequest, TimelineQuery};
use std::ffi::{CStr, CString};
use std::path::Path;
use std::ptr;
use std::slice;

// Opaque pointer to Memvid instance
pub struct MemvidHandle(Memvid);

// --- Helpers ---

fn set_error(err_out: *mut *mut c_char, err: impl ToString) {
    if !err_out.is_null() {
        let c_str = CString::new(err.to_string()).unwrap_or_default();
        unsafe { *err_out = c_str.into_raw() };
    }
}

unsafe fn ptr_to_str<'a>(ptr: *const c_char) -> Option<&'a str> {
    if ptr.is_null() {
        return None;
    }
    CStr::from_ptr(ptr).to_str().ok()
}

// --- Lifecycle ---

#[no_mangle]
pub extern "C" fn memvid_create(
    path: *const c_char,
    handle_out: *mut *mut MemvidHandle,
    err_out: *mut *mut c_char,
) -> i32 {
    let path_str = unsafe {
        match ptr_to_str(path) {
            Some(s) => s,
            None => {
                set_error(err_out, "Invalid path string");
                return -1;
            }
        }
    };

    match Memvid::create(Path::new(path_str)) {
        Ok(mem) => unsafe {
            *handle_out = Box::into_raw(Box::new(MemvidHandle(mem)));
            0
        },
        Err(e) => {
            set_error(err_out, e);
            -1
        }
    }
}

#[no_mangle]
pub extern "C" fn memvid_open(
    path: *const c_char,
    handle_out: *mut *mut MemvidHandle,
    err_out: *mut *mut c_char,
) -> i32 {
    let path_str = unsafe {
        match ptr_to_str(path) {
            Some(s) => s,
            None => {
                set_error(err_out, "Invalid path string");
                return -1;
            }
        }
    };

    match Memvid::open(Path::new(path_str)) {
        Ok(mem) => unsafe {
            *handle_out = Box::into_raw(Box::new(MemvidHandle(mem)));
            0
        },
        Err(e) => {
            set_error(err_out, e);
            -1
        }
    }
}

#[no_mangle]
pub extern "C" fn memvid_close(handle: *mut MemvidHandle) {
    if !handle.is_null() {
        unsafe {
            let _ = Box::from_raw(handle);
        }
    }
}

// --- Operations ---

#[no_mangle]
pub extern "C" fn memvid_put(
    handle: *mut MemvidHandle,
    payload: *const u8,
    payload_len: size_t,
    options_json: *const c_char,
    seq_out: *mut u64,
    err_out: *mut *mut c_char,
) -> i32 {
    let mem = unsafe { &mut (*handle).0 };
    let bytes = unsafe { slice::from_raw_parts(payload, payload_len) };
    
    let options: PutOptions = if !options_json.is_null() {
        let s = unsafe { ptr_to_str(options_json).unwrap_or("{}") };
        serde_json::from_str(s).unwrap_or_default()
    } else {
        PutOptions::default()
    };

    match mem.put_bytes_with_options(bytes, options) {
        Ok(seq) => unsafe {
            if !seq_out.is_null() {
                *seq_out = seq;
            }
            0
        },
        Err(e) => {
            set_error(err_out, e);
            -1
        }
    }
}

#[no_mangle]
pub extern "C" fn memvid_commit(
    handle: *mut MemvidHandle,
    err_out: *mut *mut c_char,
) -> i32 {
    let mem = unsafe { &mut (*handle).0 };
    match mem.commit() {
        Ok(_) => 0,
        Err(e) => {
            set_error(err_out, e);
            -1
        }
    }
}

#[no_mangle]
pub extern "C" fn memvid_search(
    handle: *mut MemvidHandle,
    request_json: *const c_char,
    response_json_out: *mut *mut c_char,
    err_out: *mut *mut c_char,
) -> i32 {
    let mem = unsafe { &mut (*handle).0 };
    let req_str = unsafe { ptr_to_str(request_json).unwrap_or("{}") };
    
    let request: SearchRequest = match serde_json::from_str(req_str) {
        Ok(r) => r,
        Err(e) => {
            set_error(err_out, format!("JSON parse error: {}", e));
            return -1;
        }
    };

    match mem.search(request) {
        Ok(response) => {
            let json = serde_json::to_string(&response).unwrap();
            let c_str = CString::new(json).unwrap();
            unsafe { *response_json_out = c_str.into_raw() };
            0
        },
        Err(e) => {
            set_error(err_out, e);
            -1
        }
    }
}

#[no_mangle]
pub extern "C" fn memvid_timeline(
    handle: *mut MemvidHandle,
    query_json: *const c_char,
    response_json_out: *mut *mut c_char,
    err_out: *mut *mut c_char,
) -> i32 {
    let mem = unsafe { &mut (*handle).0 };
    let query_str = unsafe { ptr_to_str(query_json).unwrap_or("{}") };

    let query: TimelineQuery = match serde_json::from_str(query_str) {
        Ok(q) => q,
        Err(e) => {
            set_error(err_out, format!("JSON parse error: {}", e));
            return -1;
        }
    };

    match mem.timeline(query) {
        Ok(entries) => {
            let json = serde_json::to_string(&entries).unwrap();
            let c_str = CString::new(json).unwrap();
            unsafe { *response_json_out = c_str.into_raw() };
            0
        },
        Err(e) => {
            set_error(err_out, e);
            -1
        }
    }
}

#[no_mangle]
pub extern "C" fn memvid_stats(
    handle: *mut MemvidHandle,
    response_json_out: *mut *mut c_char,
    err_out: *mut *mut c_char,
) -> i32 {
    let mem = unsafe { &(*handle).0 };
    match mem.stats() {
        Ok(stats) => {
            let json = serde_json::to_string(&stats).unwrap();
            let c_str = CString::new(json).unwrap();
            unsafe { *response_json_out = c_str.into_raw() };
            0
        },
        Err(e) => {
            set_error(err_out, e);
            -1
        }
    }
}

#[no_mangle]
pub extern "C" fn memvid_enable_lex(
    handle: *mut MemvidHandle,
    err_out: *mut *mut c_char
) -> i32 {
    let mem = unsafe { &mut (*handle).0 };
    match mem.enable_lex() {
        Ok(_) => 0,
        Err(e) => {
            set_error(err_out, e);
            -1
        }
    }
}

#[no_mangle]
pub extern "C" fn memvid_enable_vec(
    handle: *mut MemvidHandle,
    err_out: *mut *mut c_char
) -> i32 {
    let mem = unsafe { &mut (*handle).0 };
    match mem.enable_vec() {
        Ok(_) => 0,
        Err(e) => {
            set_error(err_out, e);
            -1
        }
    }
}

#[no_mangle]
pub extern "C" fn memvid_verify(
    path: *const c_char,
    deep: i32,
    response_json_out: *mut *mut c_char,
    err_out: *mut *mut c_char
) -> i32 {
    let path_str = unsafe { ptr_to_str(path).unwrap() };
    match Memvid::verify(path_str, deep != 0) {
        Ok(report) => {
             let json = serde_json::to_string(&report).unwrap();
            let c_str = CString::new(json).unwrap();
            unsafe { *response_json_out = c_str.into_raw() };
            0
        }
        Err(e) => {
            set_error(err_out, e);
            -1
        }
    }
}

// --- Utils ---

#[no_mangle]
pub extern "C" fn memvid_free_string(s: *mut c_char) {
    if !s.is_null() {
        unsafe {
            let _ = CString::from_raw(s);
        }
    }
}
