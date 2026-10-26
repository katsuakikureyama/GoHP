//go:build gohp_pcre2 && cgo

#include "bridge.h"

#include <stdlib.h>
#include <string.h>

#define GOHP_PCRE2_MAX_CAPTURES 64U
#define GOHP_PCRE2_MESSAGE_SIZE 256U
#define GOHP_PCRE2_VERSION_SIZE 64U

typedef struct {
	uint32_t number;
	char *name;
	size_t length;
} pcre2_bridge_name;

typedef struct {
	int present;
	size_t start;
	size_t end;
	unsigned char *bytes;
	size_t length;
} pcre2_bridge_capture;

struct pcre2_bridge_response {
	int status;
	int return_code;
	int error_code;
	size_t error_offset;
	char message[GOHP_PCRE2_MESSAGE_SIZE];
	char version[GOHP_PCRE2_VERSION_SIZE];
	uint32_t capture_count;
	uint32_t name_count;
	pcre2_bridge_name *names;
	uint32_t result_count;
	pcre2_bridge_capture *captures;
};

static void bridge_message(pcre2_bridge_response *response, int code) {
	PCRE2_UCHAR message[GOHP_PCRE2_MESSAGE_SIZE] = {0};
	int result = pcre2_get_error_message(code, message, sizeof(message));
	if (result >= 0) {
		(void)memcpy(response->message, message, sizeof(response->message) - 1U);
		response->message[sizeof(response->message) - 1U] = '\0';
	}
}

static unsigned char *copy_input(const unsigned char *input, size_t length) {
	unsigned char *copy = malloc(length == 0 ? 1U : length);
	if (copy == NULL) {
		return NULL;
	}
	if (length != 0) {
		(void)memcpy(copy, input, length);
	}
	return copy;
}

static int bridge_pattern_info(pcre2_code *code, uint32_t what, void *where) {
	return pcre2_pattern_info(code, what, where);
}

static int bridge_profile_matches(pcre2_code *code) {
	uint32_t options = 0;
	uint32_t newline = 0;
	uint32_t bsr = 0;
	uint32_t pattern_match_limit = 0;
	uint32_t pattern_depth_limit = 0;
	uint32_t pattern_heap_limit = 0;
	const uint32_t profile_mask = PCRE2_DOTALL | PCRE2_UTF | PCRE2_UCP;

	if (bridge_pattern_info(code, PCRE2_INFO_ALLOPTIONS, &options) != 0 ||
		bridge_pattern_info(code, PCRE2_INFO_NEWLINE, &newline) != 0 ||
		bridge_pattern_info(code, PCRE2_INFO_BSR, &bsr) != 0 ||
		bridge_pattern_info(code, PCRE2_INFO_MATCHLIMIT, &pattern_match_limit) != PCRE2_ERROR_UNSET ||
		bridge_pattern_info(code, PCRE2_INFO_DEPTHLIMIT, &pattern_depth_limit) != PCRE2_ERROR_UNSET ||
		bridge_pattern_info(code, PCRE2_INFO_HEAPLIMIT, &pattern_heap_limit) != PCRE2_ERROR_UNSET) {
		return 0;
	}
	return (options & profile_mask) == PCRE2_DOTALL && newline == PCRE2_NEWLINE_LF &&
		bsr == PCRE2_BSR_UNICODE;
}

static int bridge_copy_names(pcre2_bridge_response *response, pcre2_code *code) {
	uint32_t entry_size = 0;
	PCRE2_SPTR table = NULL;
	uint32_t i;

	if (response->name_count == 0) {
		return 1;
	}
	if (response->name_count > response->capture_count ||
		bridge_pattern_info(code, PCRE2_INFO_NAMEENTRYSIZE, &entry_size) != 0 || entry_size < 3 || entry_size > 4099 ||
		bridge_pattern_info(code, PCRE2_INFO_NAMETABLE, &table) != 0 || table == NULL) {
		return 0;
	}
	response->names = calloc(response->name_count, sizeof(*response->names));
	if (response->names == NULL) {
		return 0;
	}
	for (i = 0; i < response->name_count; i++) {
		PCRE2_SPTR entry = table + ((size_t)i * entry_size);
		size_t length = 0;
		while (length + 2U < entry_size && entry[2U + length] != '\0') {
			length++;
		}
		if (length + 2U >= entry_size) {
			return 0;
		}
		response->names[i].name = malloc(length + 1U);
		if (response->names[i].name == NULL) {
			return 0;
		}
		if (length != 0) {
			(void)memcpy(response->names[i].name, entry + 2U, length);
		}
		response->names[i].name[length] = '\0';
		response->names[i].length = length;
		response->names[i].number = ((uint32_t)entry[0] << 8U) | (uint32_t)entry[1];
		if (response->names[i].number == 0 || response->names[i].number > response->capture_count) {
			return 0;
		}
	}
	return 1;
}

pcre2_bridge_response *pcre2_bridge_match(const unsigned char *pattern, size_t pattern_len,
		const unsigned char *subject, size_t subject_len, size_t offset,
		uint32_t match_limit, uint32_t depth_limit, uint32_t heap_limit_kib) {
	pcre2_bridge_response *response = calloc(1U, sizeof(*response));
	unsigned char *pattern_copy = NULL;
	unsigned char *subject_copy = NULL;
	pcre2_compile_context *compile_context = NULL;
	pcre2_match_context *match_context = NULL;
	pcre2_match_data *match_data = NULL;
	pcre2_code *code = NULL;
	uint32_t capture_count = 0;
	int compile_error = 0;
	PCRE2_SIZE compile_offset = 0;
	int result;

	if (response == NULL) {
		return NULL;
	}
	response->status = PCRE2_BRIDGE_OK;
	if (pattern_len > 0 && pattern == NULL) {
		response->status = PCRE2_BRIDGE_ERROR_INTERNAL;
		response->return_code = PCRE2_BRIDGE_ERROR_INTERNAL;
		(void)memcpy(response->message, "missing pattern buffer", 23U);
		return response;
	}
	if (subject_len > 0 && subject == NULL) {
		response->status = PCRE2_BRIDGE_ERROR_INTERNAL;
		response->return_code = PCRE2_BRIDGE_ERROR_INTERNAL;
		(void)memcpy(response->message, "missing subject buffer", 23U);
		return response;
	}
	if (pattern_len > 4096 || subject_len > 4096 || offset > subject_len ||
		match_limit == 0 || match_limit > 100000 || depth_limit == 0 || depth_limit > 10000 ||
		heap_limit_kib == 0 || heap_limit_kib > 131072) {
		response->status = PCRE2_BRIDGE_ERROR_SIZE;
		response->return_code = PCRE2_BRIDGE_ERROR_SIZE;
		(void)memcpy(response->message, "invalid bridge bounds", 22U);
		return response;
	}

	pattern_copy = copy_input(pattern, pattern_len);
	subject_copy = copy_input(subject, subject_len);
	if (pattern_copy == NULL || subject_copy == NULL) {
		response->status = PCRE2_BRIDGE_ERROR_ALLOC;
		response->return_code = PCRE2_BRIDGE_ERROR_ALLOC;
		(void)memcpy(response->message, "cannot allocate input copy", 27U);
		goto done;
	}
	result = pcre2_config(PCRE2_CONFIG_VERSION, NULL);
	if (result <= 0 || (size_t)result > sizeof(response->version) ||
		pcre2_config(PCRE2_CONFIG_VERSION, response->version) < 0) {
		response->status = PCRE2_BRIDGE_ERROR_INTERNAL;
		response->return_code = PCRE2_BRIDGE_ERROR_INTERNAL;
		(void)memcpy(response->message, "cannot read PCRE2 version", 25U);
		goto done;
	}
	compile_context = pcre2_compile_context_create(NULL);
	if (compile_context == NULL) {
		response->status = PCRE2_BRIDGE_ERROR_ALLOC;
		response->return_code = PCRE2_BRIDGE_ERROR_ALLOC;
		(void)memcpy(response->message, "cannot create compile context", 30U);
		goto done;
	}
	/* Leaving character tables unset selects PCRE2's built-in C tables. Do not
	 * create locale-derived tables here: the profile must be process-invariant. */
	if (pcre2_set_newline(compile_context, PCRE2_NEWLINE_LF) != 0 ||
		pcre2_set_bsr(compile_context, PCRE2_BSR_UNICODE) != 0) {
		response->status = PCRE2_BRIDGE_ERROR_INTERNAL;
		response->return_code = PCRE2_BRIDGE_ERROR_INTERNAL;
		(void)memcpy(response->message, "cannot configure fixed profile", 31U);
		goto done;
	}
	code = pcre2_compile(pattern_copy, pattern_len, PCRE2_DOTALL | PCRE2_NEVER_UTF | PCRE2_NEVER_UCP,
		&compile_error, &compile_offset, compile_context);
	if (code == NULL) {
		response->status = PCRE2_BRIDGE_COMPILE;
		response->error_code = compile_error;
		response->error_offset = (size_t)compile_offset;
		bridge_message(response, compile_error);
		goto done;
	}
	if (!bridge_profile_matches(code)) {
		response->status = PCRE2_BRIDGE_ERROR_PROFILE;
		response->return_code = PCRE2_BRIDGE_ERROR_PROFILE;
		(void)memcpy(response->message, "pattern changes fixed profile", 30U);
		goto done;
	}
	if (bridge_pattern_info(code, PCRE2_INFO_CAPTURECOUNT, &capture_count) != 0 ||
		capture_count > GOHP_PCRE2_MAX_CAPTURES) {
		response->status = PCRE2_BRIDGE_ERROR_SIZE;
		response->return_code = PCRE2_BRIDGE_ERROR_SIZE;
		(void)memcpy(response->message, "capture count exceeds bridge bound", 35U);
		goto done;
	}
	response->capture_count = capture_count;
	if (bridge_pattern_info(code, PCRE2_INFO_NAMECOUNT, &response->name_count) != 0 ||
		!bridge_copy_names(response, code)) {
		response->status = PCRE2_BRIDGE_ERROR_INTERNAL;
		response->return_code = PCRE2_BRIDGE_ERROR_INTERNAL;
		(void)memcpy(response->message, "cannot copy capture names", 25U);
		goto done;
	}
	match_context = pcre2_match_context_create(NULL);
	if (match_context == NULL || pcre2_set_match_limit(match_context, match_limit) != 0 ||
		pcre2_set_depth_limit(match_context, depth_limit) != 0 ||
		pcre2_set_heap_limit(match_context, heap_limit_kib) != 0) {
		response->status = PCRE2_BRIDGE_ERROR_INTERNAL;
		response->return_code = PCRE2_BRIDGE_ERROR_INTERNAL;
		(void)memcpy(response->message, "cannot configure match limits", 29U);
		goto done;
	}
	match_data = pcre2_match_data_create_from_pattern(code, NULL);
	if (match_data == NULL || pcre2_get_ovector_count(match_data) != capture_count + 1U) {
		response->status = PCRE2_BRIDGE_ERROR_ALLOC;
		response->return_code = PCRE2_BRIDGE_ERROR_ALLOC;
		(void)memcpy(response->message, "cannot allocate match data", 27U);
		goto done;
	}
	result = pcre2_match(code, subject_copy, subject_len, offset, 0, match_data, match_context);
	response->return_code = result;
	if (result > 0) {
		PCRE2_SIZE *ovector;
		uint32_t i;
		if ((uint32_t)result > capture_count + 1U ||
			(uint32_t)result > GOHP_PCRE2_MAX_CAPTURES + 1U) {
			response->status = PCRE2_BRIDGE_ERROR_SIZE;
			response->return_code = PCRE2_BRIDGE_ERROR_SIZE;
			(void)memcpy(response->message, "match result exceeds bridge bound", 33U);
			goto done;
		}
		response->captures = calloc((uint32_t)result, sizeof(*response->captures));
		if (response->captures == NULL) {
			response->status = PCRE2_BRIDGE_ERROR_ALLOC;
			response->return_code = PCRE2_BRIDGE_ERROR_ALLOC;
			(void)memcpy(response->message, "cannot allocate capture results", 31U);
			goto done;
		}
		response->result_count = (uint32_t)result;
		ovector = pcre2_get_ovector_pointer(match_data);
		if (ovector == NULL) {
			response->status = PCRE2_BRIDGE_ERROR_INTERNAL;
			response->return_code = PCRE2_BRIDGE_ERROR_INTERNAL;
			(void)memcpy(response->message, "cannot obtain capture vector", 29U);
			goto done;
		}
		for (i = 0; i < response->result_count; i++) {
			PCRE2_SIZE start = ovector[2U * i];
			PCRE2_SIZE end = ovector[2U * i + 1U];
			if (start == PCRE2_UNSET || end == PCRE2_UNSET) {
				continue;
			}
			if (start > end || end > subject_len) {
				response->status = PCRE2_BRIDGE_ERROR_INTERNAL;
				response->return_code = PCRE2_BRIDGE_ERROR_INTERNAL;
				(void)memcpy(response->message, "invalid capture bounds", 23U);
				goto done;
			}
			response->captures[i].present = 1;
			response->captures[i].start = (size_t)start;
			response->captures[i].end = (size_t)end;
			response->captures[i].length = (size_t)(end - start);
			response->captures[i].bytes = copy_input(subject_copy + start, response->captures[i].length);
			if (response->captures[i].bytes == NULL) {
				response->status = PCRE2_BRIDGE_ERROR_ALLOC;
				response->return_code = PCRE2_BRIDGE_ERROR_ALLOC;
				(void)memcpy(response->message, "cannot copy capture bytes", 26U);
				goto done;
			}
		}
	} else if (result != PCRE2_ERROR_NOMATCH && result != PCRE2_ERROR_MATCHLIMIT) {
		bridge_message(response, result);
	}

done:
	pcre2_match_data_free(match_data);
	pcre2_match_context_free(match_context);
	pcre2_code_free(code);
	pcre2_compile_context_free(compile_context);
	free(subject_copy);
	free(pattern_copy);
	return response;
}

void pcre2_bridge_response_free(pcre2_bridge_response *response) {
	uint32_t i;
	if (response == NULL) {
		return;
	}
	for (i = 0; response->names != NULL && i < response->name_count; i++) {
		free(response->names[i].name);
	}
	for (i = 0; response->captures != NULL && i < response->result_count; i++) {
		free(response->captures[i].bytes);
	}
	free(response->captures);
	free(response->names);
	free(response);
}

int pcre2_bridge_response_status(const pcre2_bridge_response *response) { return response->status; }
int pcre2_bridge_response_return_code(const pcre2_bridge_response *response) { return response->return_code; }
int pcre2_bridge_response_error_code(const pcre2_bridge_response *response) { return response->error_code; }
size_t pcre2_bridge_response_error_offset(const pcre2_bridge_response *response) { return response->error_offset; }
const char *pcre2_bridge_response_message(const pcre2_bridge_response *response) { return response->message; }
const char *pcre2_bridge_response_version(const pcre2_bridge_response *response) { return response->version; }
uint32_t pcre2_bridge_response_capture_count(const pcre2_bridge_response *response) { return response->capture_count; }
uint32_t pcre2_bridge_response_name_count(const pcre2_bridge_response *response) { return response->name_count; }
uint32_t pcre2_bridge_response_name_number(const pcre2_bridge_response *response, uint32_t index) { return response->names[index].number; }
const char *pcre2_bridge_response_name(const pcre2_bridge_response *response, uint32_t index) { return response->names[index].name; }
size_t pcre2_bridge_response_name_length(const pcre2_bridge_response *response, uint32_t index) { return response->names[index].length; }
uint32_t pcre2_bridge_response_result_count(const pcre2_bridge_response *response) { return response->result_count; }
int pcre2_bridge_response_capture_present(const pcre2_bridge_response *response, uint32_t index) { return response->captures[index].present; }
size_t pcre2_bridge_response_capture_start(const pcre2_bridge_response *response, uint32_t index) { return response->captures[index].start; }
size_t pcre2_bridge_response_capture_end(const pcre2_bridge_response *response, uint32_t index) { return response->captures[index].end; }
const unsigned char *pcre2_bridge_response_capture_bytes(const pcre2_bridge_response *response, uint32_t index) { return response->captures[index].bytes; }
size_t pcre2_bridge_response_capture_length(const pcre2_bridge_response *response, uint32_t index) { return response->captures[index].length; }
