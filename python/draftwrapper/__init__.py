import platform

version = "v0.0.1"

system = platform.uname()[0].lower()
arch = platform.version().lower()
if "arm" in arch:
    arch = "arm64"
else:
    arch = "amd64"

binary_name = f"draftv2-{system}-{arch}"

