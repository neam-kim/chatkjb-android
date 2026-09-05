plugins {
    alias(libs.plugins.android.library)
}

android {
    namespace = "com.limelight"
    compileSdk = 37
    ndkVersion = "29.0.14206865"

    defaultConfig {
        minSdk = 26
        buildConfigField("boolean", "ROOT_BUILD", "false")
        buildConfigField("String", "APPLICATION_ID", "\"com.termux\"")
    }

    buildFeatures {
        buildConfig = true
        resValues = true
    }

    compileOptions {
        isCoreLibraryDesugaringEnabled = true
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    externalNativeBuild {
        ndkBuild {
            path = file("src/main/jni/Android.mk")
        }
    }

    packaging {
        resources.excludes += "/META-INF/*.md"
        jniLibs.useLegacyPackaging = true
    }
}

dependencies {
    implementation("org.bouncycastle:bcprov-jdk18on:1.85.2")
    implementation("org.bouncycastle:bcpkix-jdk18on:1.85")
    implementation("org.jcodec:jcodec:0.2.5")
    implementation("com.squareup.okhttp3:okhttp:5.5.0")
    implementation("org.jmdns:jmdns:3.6.3")
    implementation("com.github.cgutman:ShieldControllerExtensions:1.0.1")
    coreLibraryDesugaring("com.android.tools:desugar_jdk_libs:2.1.5")
}
