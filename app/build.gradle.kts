plugins {
    alias(libs.plugins.android.application) apply false
    alias(libs.plugins.android.library) apply false
    alias(libs.plugins.kotlin.serialization) apply false
    alias(libs.plugins.compose.compiler) apply false
}

// Imported Termux modules retain their upstream Groovy build scripts. Keep their
// project properties local to this build instead of adding a second gradle.properties.
allprojects {
    extra["minSdkVersion"] = 26
    extra["targetSdkVersion"] = 36
    extra["compileSdkVersion"] = 37
    extra["ndkVersion"] = "29.0.14206865"
    extra["markwonVersion"] = "4.6.2"
}
