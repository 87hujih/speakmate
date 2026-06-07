export function selectSupportedAudioMimeType(isTypeSupported: (candidate: string) => boolean) {
  const candidates = ["audio/ogg;codecs=opus", "audio/mp4", "audio/wav", "audio/webm;codecs=opus", "audio/webm"];

  return candidates.find((candidate) => isTypeSupported(candidate)) ?? "";
}

export function extensionForAudioMimeType(mimeType: string) {
  if (mimeType.includes("mp4")) {
    return "m4a";
  }
  if (mimeType.includes("ogg")) {
    return "ogg";
  }
  if (mimeType.includes("wav")) {
    return "wav";
  }

  return "webm";
}
