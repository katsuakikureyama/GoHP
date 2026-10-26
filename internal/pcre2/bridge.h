#ifndef GOHP_PCRE2_BRIDGE_H
#define GOHP_PCRE2_BRIDGE_H

#include <stddef.h>
#include <stdint.h>

#define PCRE2_CODE_UNIT_WIDTH 8
#include <pcre2.h>

enum {
	PCRE2_BRIDGE_OK = 0,
	PCRE2_BRIDGE_COMPILE = 1,
	PCRE2_BRIDGE_ERROR_ALLOC = -10000,
	PCRE2_BRIDGE_ERROR_PROFILE = -10001,
	PCRE2_BRIDGE_ERROR_SIZE = -10002,
	PCRE2_BRIDGE_ERROR_INTERNAL = -10003
};

typedef struct pcre2_bridge_response pcre2_bridge_response;

pcre2_bridge_response *pcre2_bridge_match(const unsigned char *pattern, size_t pattern_len,
		const unsigned char *subject, size_t subject_len, size_t offset,
		uint32_t match_limit, uint32_t depth_limit, uint32_t heap_limit_kib);
void pcre2_bridge_response_free(pcre2_bridge_response *response);

int pcre2_bridge_response_status(const pcre2_bridge_response *response);
int pcre2_bridge_response_return_code(const pcre2_bridge_response *response);
int pcre2_bridge_response_error_code(const pcre2_bridge_response *response);
size_t pcre2_bridge_response_error_offset(const pcre2_bridge_response *response);
const char *pcre2_bridge_response_message(const pcre2_bridge_response *response);
const char *pcre2_bridge_response_version(const pcre2_bridge_response *response);
uint32_t pcre2_bridge_response_capture_count(const pcre2_bridge_response *response);
uint32_t pcre2_bridge_response_name_count(const pcre2_bridge_response *response);
uint32_t pcre2_bridge_response_name_number(const pcre2_bridge_response *response, uint32_t index);
const char *pcre2_bridge_response_name(const pcre2_bridge_response *response, uint32_t index);
size_t pcre2_bridge_response_name_length(const pcre2_bridge_response *response, uint32_t index);
uint32_t pcre2_bridge_response_result_count(const pcre2_bridge_response *response);
int pcre2_bridge_response_capture_present(const pcre2_bridge_response *response, uint32_t index);
size_t pcre2_bridge_response_capture_start(const pcre2_bridge_response *response, uint32_t index);
size_t pcre2_bridge_response_capture_end(const pcre2_bridge_response *response, uint32_t index);
const unsigned char *pcre2_bridge_response_capture_bytes(const pcre2_bridge_response *response,
		uint32_t index);
size_t pcre2_bridge_response_capture_length(const pcre2_bridge_response *response, uint32_t index);

#endif
