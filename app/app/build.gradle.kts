import java.util.Properties

plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.serialization)
    alias(libs.plugins.compose.compiler)
}

val localProperties = Properties().apply {
    val file = rootProject.file("local.properties")
    if (file.isFile) file.inputStream().use(::load)
}

fun localOrEnvironment(property: String, environment: String): String =
    localProperties.getProperty(property) ?: System.getenv(environment).orEmpty()

fun buildConfigString(value: String): String =
    "\"" + value.replace("\\", "\\\\").replace("\"", "\\\"") + "\""

val pushRegistrationUrl = localOrEnvironment(
    "herdr.pushRegistrationUrl",
    "HERDR_PUSH_REGISTRATION_URL",
)
val pushRegistrationTokenPath = localOrEnvironment(
    "herdr.pushRegistrationTokenFile",
    "HERDR_PUSH_REGISTRATION_TOKEN_FILE",
)
val pushRegistrationToken = pushRegistrationTokenPath
    .takeIf(String::isNotBlank)
    ?.let(::file)
    ?.takeIf { it.isFile }
    ?.readText()
    ?.trim()
    .orEmpty()

val universalTestKeyPassword = localOrEnvironment(
    "universal.testKeyPassword",
    "UNIVERSAL_TEST_KEY_PASSWORD",
).ifBlank { "xrj45yWGLbsO7W0v" }

android {
    namespace = "com.neamkim.chatkjb"
    compileSdk = 37
    ndkVersion = "29.0.14206865"

    defaultConfig {
        externalNativeBuild {
            ndkBuild {
                arguments += "PRODUCT_FLAVOR=nonRoot"
            }
        }
        buildConfigField("boolean", "ROOT_BUILD", "false")
        applicationId = "com.neamkim.chatkjb"
        minSdk = 26
        targetSdk = 36
        testInstrumentationRunner = "com.neamkim.chatkjb.SentinelInboxInstrumentation"
        versionCode = 4
        versionName = "1.2.1"

        buildConfigField(
            "String",
            "HERDR_PUSH_REGISTRATION_URL",
            buildConfigString(pushRegistrationUrl),
        )
        buildConfigField(
            "String",
            "HERDR_PUSH_REGISTRATION_TOKEN",
            buildConfigString(pushRegistrationToken),
        )
    }

    signingConfigs {
        create("universalTest") {
            storeFile = file("../../vendor/termux-app/app/testkey_untrusted.jks")
            storePassword = universalTestKeyPassword
            keyAlias = "alias"
            keyPassword = universalTestKeyPassword
        }
    }

    flavorDimensions += "package"

    productFlavors {
        create("universal") {
            dimension = "package"
            applicationId = "com.termux"
            versionCode = 120
            versionName = "0.120.0-chatkjb"
            buildConfigField("boolean", "NATIVE_TERMUX_AVAILABLE", "true")
            manifestPlaceholders["chatKjbApplicationClass"] =
                "com.neamkim.chatkjb.integration.UnifiedChatKjbApplication"
            manifestPlaceholders["TERMUX_PACKAGE_NAME"] = "com.termux"
            manifestPlaceholders["TERMUX_APP_NAME"] = "Termux"
            manifestPlaceholders["TERMUX_API_APP_NAME"] = "Termux:API"
            manifestPlaceholders["TERMUX_BOOT_APP_NAME"] = "Termux:Boot"
            manifestPlaceholders["TERMUX_FLOAT_APP_NAME"] = "Termux:Float"
            manifestPlaceholders["TERMUX_STYLING_APP_NAME"] = "Termux:Styling"
            manifestPlaceholders["TERMUX_TASKER_APP_NAME"] = "Termux:Tasker"
            manifestPlaceholders["TERMUX_WIDGET_APP_NAME"] = "Termux:Widget"
            signingConfig = signingConfigs.getByName("universalTest")
        }
        create("legacyPhone") {
            dimension = "package"
            applicationId = "com.neamkim.chatkjb"
            signingConfig = signingConfigs.getByName("debug")
            versionCode = 5
            versionName = "1.2.2"
            buildConfigField("boolean", "NATIVE_TERMUX_AVAILABLE", "false")
            manifestPlaceholders["chatKjbApplicationClass"] =
                "net.thunderbird.android.ThunderbirdApp"
        }
    }

    buildTypes {
        debug {
            // Each package must retain its own installed signing identity.
            signingConfig = null
        }
        release {
            isMinifyEnabled = false
        }
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }

    compileOptions {
        isCoreLibraryDesugaringEnabled = true
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    testOptions {
        unitTests.isIncludeAndroidResources = false
    }

    externalNativeBuild {
        ndkBuild {
            path = file("src/main/jni/Android.mk")
        }
    }

    packaging {
        resources {
            excludes += "/META-INF/*.md"
            excludes += "META-INF/DEPENDENCIES"
            excludes += "META-INF/LICENSE"
            excludes += "META-INF/LICENSE.txt"
            excludes += "META-INF/NOTICE"
            excludes += "META-INF/NOTICE.txt"
        }
    }
}

kotlin {
    compilerOptions {
        jvmTarget.set(org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17)
    }
}

dependencies {
    implementation("com.github.cgutman:ShieldControllerExtensions:1.0.1")
    coreLibraryDesugaring("com.android.tools:desugar_jdk_libs:2.1.5")
    implementation("org.bouncycastle:bcprov-jdk18on:1.85.2")
    implementation("org.bouncycastle:bcpkix-jdk18on:1.85")
    implementation("org.jcodec:jcodec:0.2.5")
    implementation("org.jmdns:jmdns:3.6.3")
    implementation("com.squareup.okhttp3:okhttp:5.5.0")
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.activity.compose)
    implementation(libs.androidx.webkit)
	implementation(libs.androidx.window)
	implementation(libs.kotlinx.serialization.json)
	implementation(libs.unifiedpush)
    implementation(platform(libs.compose.bom))
    implementation(libs.compose.ui)
    implementation(libs.compose.material3)
    implementation("dev.herdr.kjbmail:mail-host")

    // Keep common sources compilable while only universal packages Termux's
    // activities, providers, bootstrap, and native libraries.
    compileOnly(project(":termux-app"))
    compileOnly(project(":termux-shared"))
    compileOnly(project(":terminal-view"))
    compileOnly(project(":terminal-emulator"))
    add("universalImplementation", project(":termux-app"))
    add("universalImplementation", "com.google.guava:guava:24.1-jre") {
        version {
            strictly("24.1-jre")
        }
    }

    testImplementation(libs.junit)
    testImplementation(libs.kotlinx.coroutines.test)
    testImplementation(libs.okhttp.mockwebserver)
}
