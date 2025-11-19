#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <unistd.h>
#include <zlib-ng.h>

int compress_file_zlib(const char *input_path, const char *output_path)
{
	int fd_in = -1, fd_out = -1;
	void *mapped_data_in = MAP_FAILED;
	void *mapped_data_out = MAP_FAILED;

	fd_in = open(input_path, O_RDONLY);
	if (fd_in == -1) {
		perror("open input");
		goto error;
	}

	struct stat sb;
	if (fstat(fd_in, &sb) == -1) {
		perror("fstat");
		goto error;
	}
	size_t file_size = sb.st_size;

	if (file_size == 0) {
		fprintf(stderr, "File is empty\n");
		goto error;
	}

	mapped_data_in = mmap(NULL, file_size, PROT_READ, MAP_PRIVATE, fd_in, 0);
	if (mapped_data_in == MAP_FAILED) {
		perror("mmap input");
		goto error;
	}

	size_t compressed_size_bound = zng_compressBound(file_size);

	fd_out = open(output_path, O_RDWR | O_CREAT | O_TRUNC, 0644);
	if (fd_out == -1) {
		perror("open output");
		goto error;
	}

	if (ftruncate(fd_out, compressed_size_bound) == -1) {
		perror("ftruncate");
		goto error;
	}

	mapped_data_out = mmap(NULL, compressed_size_bound, PROT_READ | PROT_WRITE, 
	                      MAP_SHARED, fd_out, 0);
	if (mapped_data_out == MAP_FAILED) {
		perror("mmap output");
		goto error;
	}

	int ret = zng_compress2(mapped_data_out, &compressed_size_bound,
		(const unsigned char *)mapped_data_in, file_size,
		Z_BEST_COMPRESSION);

	if (ret != Z_OK) {
		fprintf(stderr, "Compression failed: %d\n", ret);
		goto error;
	}

	if (ftruncate(fd_out, compressed_size_bound) == -1) {
		perror("ftruncate final");
		goto error;
	}

	if (msync(mapped_data_out, compressed_size_bound, MS_SYNC) == -1) {
		perror("msync");
		goto error;
	}

	fprintf(stderr,
		"Compressed: %s (%zu bytes) -> %s (%zu bytes), ratio: %.1f%%\n",
		input_path, file_size, output_path, compressed_size_bound,
		(compressed_size_bound * 100.0) / file_size);

	munmap(mapped_data_in, file_size);
	munmap(mapped_data_out, compressed_size_bound);
	close(fd_in);
	close(fd_out);
	return 0;

error:
	if (mapped_data_in != MAP_FAILED)
		munmap(mapped_data_in, file_size);
	if (mapped_data_out != MAP_FAILED)
		munmap(mapped_data_out, compressed_size_bound);
	if (fd_in != -1)
		close(fd_in);
	if (fd_out != -1)
		close(fd_out);
	return -1;
}

int compress_files(const char **input_files, const char **output_files,
	int count)
{
	int success_count = 0;

	for (int i = 0; i < count; i++) {
		if (compress_file_zlib(input_files[i], output_files[i]) == 0) {
			success_count++;
		} else {
			fprintf(stderr, "Failed to compress: %s\n", input_files[i]);
		}
	}

	fprintf(stderr, "Successfully compressed %d/%d files\n", success_count,
		count);
	return success_count;
}

int main(int argc, char **argv)
{
	if (argc < 3 || argc % 2 != 1) {
		fprintf(stderr,
			"Usage: %s <input1> <output1.zlib> [input2 output2.zlib ...]\n",
			argv[0]);
		return 1;
	}

	int file_count = (argc - 1) / 2;
	const char **inputs = (const char **)&argv[1];
	const char **outputs = (const char **)&argv[1 + file_count];

	return compress_files(inputs, outputs, file_count) == file_count ? 0 : 1;
}
