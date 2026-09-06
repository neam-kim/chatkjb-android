pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}

dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
        maven(url = "https://jitpack.io") {
            mavenContent {
                includeGroup("com.github.ByteHamster")
                includeGroup("com.github.cketti")
                includeGroup("com.github.cgutman")
                includeGroup("com.termux")
            }
        }
    }
}

rootProject.name = "ChatKJB"

// Mail sources are owned by this repository. Never fall back to a sibling checkout.
includeBuild("../KJBMail") {
    dependencySubstitution {
        substitute(module("dev.herdr.kjbmail:mail-host")).using(project(":mail-host"))
    }
}


include(":app")
include(":termux-app")
include(":termux-shared")
include(":terminal-emulator")
include(":terminal-view")

project(":termux-app").projectDir = file("../vendor/termux-app/app")
project(":termux-shared").projectDir = file("../vendor/termux-app/termux-shared")
project(":terminal-emulator").projectDir = file("../vendor/termux-app/terminal-emulator")
project(":terminal-view").projectDir = file("../vendor/termux-app/terminal-view")
