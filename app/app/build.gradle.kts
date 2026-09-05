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
