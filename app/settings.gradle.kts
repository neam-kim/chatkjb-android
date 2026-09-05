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
