group = "dev.herdr.kjbmail"

plugins {
    id(ThunderbirdPlugins.Library.androidCompose)
    alias(libs.plugins.kotlin.serialization)
}

android {
    namespace = "net.thunderbird.android"

    buildFeatures {
        buildConfig = true
    }

    defaultConfig {
        buildConfigField("String", "CLIENT_INFO_APP_NAME", "\"KJBMail\"")
        buildConfigField("String", "GLEAN_RELEASE_CHANNEL", "null")
        buildConfigField("int", "VERSION_CODE", "4")
        buildConfigField("String", "VERSION_NAME", "\"23.0\"")
    }
}

dependencies {
    implementation(projects.appCommon)
    implementation(projects.core.ui.compose.common)
    implementation(projects.core.ui.legacy.theme2.thunderbird)
    implementation(projects.feature.launcher)

    implementation(projects.legacy.core)
    implementation(projects.legacy.ui.legacy)
    implementation(projects.core.featureflag)

    implementation(projects.feature.account.settings.impl)
    implementation(projects.feature.mail.message.list.api)
    implementation(projects.feature.mail.message.list.internal)
    implementation(projects.feature.mail.message.reader.api)

    implementation(projects.feature.widget.messageList)
    implementation(projects.feature.widget.messageListGlance)
    implementation(projects.feature.widget.shortcut)
    implementation(projects.feature.widget.unread)

    implementation(projects.feature.telemetry.noop)
    implementation(projects.feature.autodiscovery.api)
    implementation(projects.feature.funding.link)
    implementation(projects.feature.onboarding.migration.thunderbird)
    implementation(projects.feature.migration.launcher.thunderbird)
    implementation(projects.feature.thundermail.api)
    implementation(projects.feature.thundermail.thunderbird)
    implementation(libs.androidx.work.runtime)

    debugImplementation(projects.backend.demo)
    debugImplementation(projects.feature.autodiscovery.demo)
}
