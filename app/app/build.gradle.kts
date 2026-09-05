plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.serialization)
    alias(libs.plugins.compose.compiler)
}

android {
    namespace = "dev.herdr.mobile"
    compileSdk = 37
    ndkVersion = "29.0.14206865"

    defaultConfig {
        applicationId = "com.termux"
        minSdk = 26
        targetSdk = 36
        versionCode = 119
        versionName = "0.119.0-chatkjb-tab"
        manifestPlaceholders["TERMUX_PACKAGE_NAME"] = "com.termux"
        manifestPlaceholders["TERMUX_APP_NAME"] = "Termux"
        manifestPlaceholders["TERMUX_API_APP_NAME"] = "Termux:API"
        manifestPlaceholders["TERMUX_BOOT_APP_NAME"] = "Termux:Boot"
        manifestPlaceholders["TERMUX_FLOAT_APP_NAME"] = "Termux:Float"
        manifestPlaceholders["TERMUX_STYLING_APP_NAME"] = "Termux:Styling"
        manifestPlaceholders["TERMUX_TASKER_APP_NAME"] = "Termux:Tasker"
        manifestPlaceholders["TERMUX_WIDGET_APP_NAME"] = "Termux:Widget"
        ndk {
            abiFilters += listOf("arm64-v8a")
        }
    }

    signingConfigs {
        create("termuxDebug") {
            storeFile = file("../../vendor/termux-app/app/testkey_untrusted.jks")
            storePassword = "xrj45yWGLbsO7W0v"
            keyAlias = "alias"
            keyPassword = "xrj45yWGLbsO7W0v"
        }
    }

    buildTypes {
        debug {
            signingConfig = signingConfigs.getByName("termuxDebug")
        }
        release {
            isMinifyEnabled = false
            signingConfig = signingConfigs.getByName("termuxDebug")
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

    packaging {
        resources {
            excludes += "/META-INF/*.md"
            excludes += "META-INF/DEPENDENCIES"
            excludes += "META-INF/LICENSE"
            excludes += "META-INF/LICENSE.txt"
            excludes += "META-INF/NOTICE"
            excludes += "META-INF/NOTICE.txt"
        }
        jniLibs {
            useLegacyPackaging = true
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
    implementation(libs.androidx.browser)
    implementation(libs.androidx.lifecycle.viewmodel.compose)
    implementation(libs.androidx.lifecycle.runtime.compose)
    implementation(platform(libs.compose.bom))
    implementation(libs.compose.ui)
    implementation(libs.compose.material3)
    implementation(libs.compose.material.icons)
    implementation(libs.okhttp)
    implementation(libs.kotlinx.serialization.json)
    implementation(libs.kotlinx.coroutines.core)
    implementation(libs.datastore.preferences)
    implementation(libs.unifiedpush)
    implementation(project(":termux-app"))
    implementation(project(":moonlight"))
    implementation("dev.herdr.kjbmail:mail-host")
    implementation("com.google.guava:guava:24.1-jre") {
        version { strictly("24.1-jre") }
    }

    coreLibraryDesugaring("com.android.tools:desugar_jdk_libs:2.1.5")
    implementation("org.bouncycastle:bcprov-jdk18on:1.85.2")
    implementation("org.bouncycastle:bcpkix-jdk18on:1.85")
    implementation("org.jcodec:jcodec:0.2.5")
    implementation("com.squareup.okhttp3:okhttp:5.5.0")
    implementation("org.jmdns:jmdns:3.6.3")
    implementation("com.github.cgutman:ShieldControllerExtensions:1.0.1")
    testImplementation(libs.junit)
    testImplementation(libs.kotlinx.coroutines.test)
    testImplementation(libs.okhttp.mockwebserver)
}
