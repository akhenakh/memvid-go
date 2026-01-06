#ifndef MEMVID_H
#define MEMVID_H

#include <stdint.h>
#include <stddef.h>

typedef struct MemvidHandle MemvidHandle;

// Lifecycle
int memvid_create(const char* path, MemvidHandle** handle, char** error);
int memvid_open(const char* path, MemvidHandle** handle, char** error);
void memvid_close(MemvidHandle* handle);

// Operations
int memvid_put(MemvidHandle* handle, const uint8_t* payload, size_t len, const char* options_json, uint64_t* seq_out, char** error);
int memvid_commit(MemvidHandle* handle, char** error);
int memvid_search(MemvidHandle* handle, const char* request_json, char** response_json, char** error);
int memvid_timeline(MemvidHandle* handle, const char* query_json, char** response_json, char** error);
int memvid_stats(MemvidHandle* handle, char** response_json, char** error);

// Features
int memvid_enable_lex(MemvidHandle* handle, char** error);
int memvid_enable_vec(MemvidHandle* handle, char** error);

// Static
int memvid_verify(const char* path, int deep, char** response_json, char** error);

// Utility
void memvid_free_string(char* s);

#endif
