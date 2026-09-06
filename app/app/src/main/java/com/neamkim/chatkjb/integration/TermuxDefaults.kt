package com.neamkim.chatkjb.integration

import java.io.File
import java.io.InputStream
import java.nio.file.FileAlreadyExistsException
import java.nio.file.Files
import java.nio.file.StandardOpenOption

/** Install a default without replacing an existing user-selected font. */
internal fun installFontIfMissing(fontFile: File, openBundledFont: () -> InputStream) {
    if (fontFile.exists()) return
    fontFile.parentFile?.mkdirs()
    openBundledFont().use { input ->
        try {
            Files.newOutputStream(
                fontFile.toPath(), StandardOpenOption.CREATE_NEW, StandardOpenOption.WRITE,
            ).use { output -> input.copyTo(output) }
        } catch (_: FileAlreadyExistsException) {
            // Another process or the user installed a font after the initial check.
        }
    }
}
