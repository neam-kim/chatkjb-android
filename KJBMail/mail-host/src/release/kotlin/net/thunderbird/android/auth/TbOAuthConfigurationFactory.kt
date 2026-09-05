package net.thunderbird.android.auth

import net.thunderbird.core.common.oauth.OAuthConfiguration
import net.thunderbird.core.common.oauth.OAuthConfigurationFactory

private const val REDIRECT_SCHEME = "dev.herdr.mobile"

@Suppress("ktlint:standard:max-line-length")
class TbOAuthConfigurationFactory : OAuthConfigurationFactory {
    override fun createConfigurations(): Map<List<String>, OAuthConfiguration> = mapOf(
        createAolConfiguration(),
        createFastmailConfiguration(),
        createGmailConfiguration(),
        createMicrosoftConfiguration(),
        createYahooConfiguration(),
        createThundermailConfiguration(),
        createThundermailStageConfiguration(),
    )

    private fun createAolConfiguration() = listOf("imap.aol.com", "smtp.aol.com") to OAuthConfiguration(
        clientId = "dj0yJmk9MVJGcHpSejNUcTU3JmQ9WVdrOWMwMHhjSFZqTkhRbWNHbzlNQT09JnM9Y29uc3VtZXJzZWNyZXQmc3Y9MCZ4PWNk",
        scopes = listOf("mail-w"),
        authorizationEndpoint = "https://api.login.aol.com/oauth2/request_auth",
        tokenEndpoint = "https://api.login.aol.com/oauth2/get_token",
        redirectUri = "$REDIRECT_SCHEME://oauth2redirect",
    )

    private fun createFastmailConfiguration() = listOf("imap.fastmail.com", "smtp.fastmail.com") to OAuthConfiguration(
        clientId = "353e41ae",
        scopes = listOf("https://www.fastmail.com/dev/protocol-imap", "https://www.fastmail.com/dev/protocol-smtp"),
        authorizationEndpoint = "https://api.fastmail.com/oauth/authorize",
        tokenEndpoint = "https://api.fastmail.com/oauth/refresh",
        redirectUri = "$REDIRECT_SCHEME://oauth2redirect",
    )

    private fun createGmailConfiguration() =
        listOf("imap.gmail.com", "imap.googlemail.com", "smtp.gmail.com", "smtp.googlemail.com") to OAuthConfiguration(
            clientId = "560629489500-no2mlau7e4vn3psh5esaiodgri09jrj9.apps.googleusercontent.com",
            scopes = listOf("https://mail.google.com/"),
            authorizationEndpoint = "https://accounts.google.com/o/oauth2/v2/auth",
            tokenEndpoint = "https://oauth2.googleapis.com/token",
            redirectUri = "$REDIRECT_SCHEME:/oauth2redirect",
        )

    private fun createMicrosoftConfiguration() =
        listOf("outlook.office365.com", "smtp.office365.com", "smtp-mail.outlook.com") to OAuthConfiguration(
            clientId = "e6f8716e-299d-4ed9-bbf3-453f192f44e5",
            scopes = listOf(
                "profile",
                "openid",
                "email",
                "https://outlook.office.com/IMAP.AccessAsUser.All",
                "https://outlook.office.com/SMTP.Send",
                "offline_access",
            ),
            authorizationEndpoint = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
            tokenEndpoint = "https://login.microsoftonline.com/common/oauth2/v2.0/token",
            redirectUri = "msauth://dev.herdr.mobile/eaXDuh6T3KFWjcJhsoaObT9OayU%3D",
        )

    private fun createYahooConfiguration() = listOf("imap.mail.yahoo.com", "smtp.mail.yahoo.com") to OAuthConfiguration(
        clientId = "dj0yJmk9bXRhTkZod2xmY3JrJmQ9WVdrOVUyUTRXRGQ0Tlc4bWNHbzlNQT09JnM9Y29uc3VtZXJzZWNyZXQmc3Y9MCZ4PTkx",
        scopes = listOf("mail-w"),
        authorizationEndpoint = "https://api.login.yahoo.com/oauth2/request_auth",
        tokenEndpoint = "https://api.login.yahoo.com/oauth2/get_token",
        redirectUri = "$REDIRECT_SCHEME://oauth2redirect",
    )

    private fun createThundermailConfiguration() = listOf("mail.tb.pro", "mail.thundermail.com") to OAuthConfiguration(
        clientId = "mobile-android-thunderbird",
        scopes = listOf("openid", "profile", "email", "offline_access"),
        authorizationEndpoint = "https://auth.tb.pro/realms/tbpro/protocol/openid-connect/auth",
        tokenEndpoint = "https://auth.tb.pro/realms/tbpro/protocol/openid-connect/token",
        redirectUri = "$REDIRECT_SCHEME://oauth2redirect",
    )

    private fun createThundermailStageConfiguration() = listOf("mail.stage-thundermail.com") to OAuthConfiguration(
        clientId = "mobile-android-thunderbird",
        scopes = listOf("openid", "profile", "email", "offline_access"),
        authorizationEndpoint = "https://auth-stage.tb.pro/realms/tbpro/protocol/openid-connect/auth",
        tokenEndpoint = "https://auth-stage.tb.pro/realms/tbpro/protocol/openid-connect/token",
        redirectUri = "$REDIRECT_SCHEME://oauth2redirect",
    )
}
