package com.neamkim.chatkjb

import com.neamkim.chatkjb.integration.installFontIfMissing
import org.junit.Assert.assertEquals
import org.junit.Rule
import org.junit.Test
import org.junit.rules.TemporaryFolder

class TermuxDefaultsTest {
    @get:Rule val temporaryFolder = TemporaryFolder()

    @Test fun existingUserFontIsNotReplacedOrReadFromAssets() {
        val font = temporaryFolder.newFile("font.ttf").apply { writeText("user font") }
        installFontIfMissing(font) { error("Existing font must skip bundled asset access") }
        assertEquals("user font", font.readText())
    }

    @Test fun missingFontReceivesBundledDefaultOnlyOnce() {
        val font = temporaryFolder.root.resolve("home/.termux/font.ttf")
        installFontIfMissing(font) { "bundled font".byteInputStream() }
        font.writeText("later customization")
        installFontIfMissing(font) { "bundled font".byteInputStream() }
        assertEquals("later customization", font.readText())
    }
}
