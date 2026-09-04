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

android {
    namespace = "com.neamkim.chatkjb"
    compileSdk = 36

    defaultConfig {
        applicationId = "com.neamkim.chatkjb"
        minSdk = 26
        targetSdk = 36
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

    buildTypes {
        release {
            isMinifyEnabled = false
        }
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    testOptions {
        unitTests.isIncludeAndroidResources = false
    }

    packaging {
        resources {
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
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.activity.compose)
    implementation(libs.androidx.webkit)
	implementation(libs.kotlinx.serialization.json)
	implementation(libs.unifiedpush)
    implementation(platform(libs.compose.bom))
    implementation(libs.compose.ui)
    implementation(libs.compose.material3)
    implementation("dev.herdr.kjbmail:mail-host")

    testImplementation(libs.junit)
    testImplementation(libs.kotlinx.coroutines.test)
    testImplementation(libs.okhttp.mockwebserver)
}
