/** selectSupportedAudioMimeType 从候选类型中选择浏览器支持的录音 MIME。 */
export function selectSupportedAudioMimeType(isTypeSupported: (candidate: string) => boolean) {
  const candidates = ["audio/ogg;codecs=opus", "audio/mp4", "audio/wav", "audio/webm;codecs=opus", "audio/webm"];

  return candidates.find((candidate) => isTypeSupported(candidate)) ?? "";
}

/** extensionForAudioMimeType 根据音频 MIME 类型推导文件扩展名。 */
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
