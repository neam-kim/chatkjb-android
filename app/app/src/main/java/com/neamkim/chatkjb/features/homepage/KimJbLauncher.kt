package com.neamkim.chatkjb.features.homepage

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.Image
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.role
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.neamkim.chatkjb.R
import com.neamkim.chatkjb.core.navigation.AppDestination
import com.neamkim.chatkjb.features.herdr.push.AutomationInbox
import com.neamkim.chatkjb.features.herdr.push.AutomationInboxItem

internal val KimJbBackground = Color(0xFFF7F7F7)
private val KimJbInk = Color(0xFF111111)
private val KimJbMuted = Color(0xFF555555)
private val KimJbBorder = Color(0xFFDDDDDD)

private val KimJbColorScheme = lightColorScheme(
    background = KimJbBackground,
    onBackground = KimJbInk,
    surface = Color.White,
    onSurface = KimJbInk,
    surfaceVariant = Color.White,
    onSurfaceVariant = KimJbMuted,
    primary = KimJbInk,
    onPrimary = Color.White,
    outline = KimJbBorder,
    error = Color(0xFFB3261E),
)

/** Ordered launcher entries for deterministic navigation and verification. */
data class LauncherEntry(
    val title: String,
    val description: String,
    val destination: AppDestination,
)

val KimJbLauncherEntries: List<LauncherEntry> = listOf(
    LauncherEntry("Site", "Open Site", AppDestination.HOMEPAGE),
    LauncherEntry("Email", "Open Email", AppDestination.EMAIL),
    LauncherEntry("Finance", "Open Finance", AppDestination.FINANCE),
    LauncherEntry("ChatKJB", "Open ChatKJB", AppDestination.CHAT_KJB),
    LauncherEntry("Server", "Open Moonlight Server", AppDestination.MOONLIGHT),
)

val KimJbConsoleEntries: List<LauncherEntry> = listOf(
    LauncherEntry("AutoBot", "Open AutoBot console", AppDestination.AUTOBOT),
    LauncherEntry("Server", "Open Server console", AppDestination.SERVER),
)

/**
 * Native launcher styled after kimjb.com's current homepage: light canvas,
 * centered monogram, restrained typography, and bordered white link rows.
 * The ChatKJB row opens the agent UI bundled in this application.
 */
@Composable
fun KimJbLauncher(
    emailError: Boolean,
    onHomepage: () -> Unit,
    onFinance: () -> Unit,
    onEmail: () -> Unit,
    onChat: () -> Unit,
    onServer: () -> Unit,
    onConsoleSettings: () -> Unit,
) {
    MaterialTheme(colorScheme = KimJbColorScheme) {
        Surface(
            modifier = Modifier.fillMaxSize(),
            color = KimJbBackground,
        ) {
            BoxWithConstraints(modifier = Modifier.fillMaxSize()) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .verticalScroll(rememberScrollState())
                        .heightIn(min = maxHeight)
                        .padding(horizontal = 24.dp, vertical = 28.dp),
                    horizontalAlignment = androidx.compose.ui.Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center,
                ) {
                    Image(
                        painter = painterResource(R.drawable.kimjb_logo),
                        contentDescription = "Kim JongBeom logo",
                        modifier = Modifier
                            .size(72.dp)
                            .clickable(onClick = onConsoleSettings)
                            .semantics {
                                role = Role.Button
                                contentDescription = "Open console settings"
                            },
                    )
                    Spacer(Modifier.height(12.dp))
                    Text(
                        "Kim JongBeom",
                        color = KimJbInk,
                        fontSize = 28.sp,
                        fontWeight = FontWeight.Bold,
                        textAlign = TextAlign.Center,
                    )
                    Spacer(Modifier.height(20.dp))
                    Column(modifier = Modifier.fillMaxWidth()) {
                        KimJbLink(title = "Site", onClick = onHomepage, description = "Open Site")
                        Spacer(Modifier.height(12.dp))
                        KimJbLink(title = "Email", onClick = onEmail, description = "Open Email")
                        if (emailError) {
                            Spacer(Modifier.height(8.dp))
                            Text(
                                "KJB Mail 앱을 찾을 수 없습니다. 통합 메일 이전이 끝날 때까지 기존 앱을 유지해 주세요.",
                                color = KimJbColorScheme.error,
                                fontSize = 13.sp,
                                lineHeight = 19.sp,
                                modifier = Modifier.fillMaxWidth(),
                            )
                        }
                        Spacer(Modifier.height(12.dp))
                        KimJbLink(title = "Finance", onClick = onFinance, description = "Open Finance")
                        Spacer(Modifier.height(12.dp))
                        KimJbLink(title = "ChatKJB", onClick = onChat, description = "Open ChatKJB")
                        Spacer(Modifier.height(12.dp))
                        KimJbLink(title = "Server", onClick = onServer, description = "Open Moonlight Server")
                    }
                }
            }
        }
    }
}

/** Hidden console-settings destination reached only by tapping the launcher logo. */
@Composable
fun KimJbConsoleSettings(
    onAutoBot: () -> Unit,
    onServer: () -> Unit,
) {
    val context = LocalContext.current
    var skill by remember { mutableStateOf(AutomationInbox.skill(context)) }
    var sentinel by remember { mutableStateOf(AutomationInbox.sentinel(context)) }
    DisposableEffect(context) {
        val preferences = AutomationInbox.preferences(context)
        val listener = android.content.SharedPreferences.OnSharedPreferenceChangeListener { _, _ ->
            skill = AutomationInbox.skill(context)
            sentinel = AutomationInbox.sentinel(context)
        }
        preferences.registerOnSharedPreferenceChangeListener(listener)
        onDispose { preferences.unregisterOnSharedPreferenceChangeListener(listener) }
    }
    MaterialTheme(colorScheme = KimJbColorScheme) {
        Surface(
            modifier = Modifier.fillMaxSize(),
            color = KimJbBackground,
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .verticalScroll(rememberScrollState())
                    .padding(horizontal = 24.dp, vertical = 28.dp),
                horizontalAlignment = androidx.compose.ui.Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center,
            ) {
                Text(
                    "Command Center",
                    color = KimJbInk,
                    fontSize = 26.sp,
                    fontWeight = FontWeight.Bold,
                )
                Spacer(Modifier.height(20.dp))
                KimJbLink(
                    title = "AutoBot",
                    onClick = onAutoBot,
                    description = "Open AutoBot console",
                )
                Spacer(Modifier.height(12.dp))
                KimJbLink(
                    title = "Server",
                    onClick = onServer,
                    description = "Open Server console",
                )
                Spacer(Modifier.height(28.dp))
                AutomationSection(
                    title = "Skill Suggestions",
                    emptyText = "새로운 제안이 없습니다.",
                    item = skill,
                    isProblem = false,
                )
                Spacer(Modifier.height(16.dp))
                AutomationSection(
                    title = "Sentinel",
                    emptyText = "감지된 문제가 없습니다.",
                    item = sentinel,
                    isProblem = sentinel != null,
                )
            }
        }
    }
}

@Composable
private fun AutomationSection(
    title: String,
    emptyText: String,
    item: AutomationInboxItem?,
    isProblem: Boolean,
) {
    Column(modifier = Modifier.fillMaxWidth()) {
        Text(
            title,
            color = KimJbInk,
            fontSize = 18.sp,
            fontWeight = FontWeight.Bold,
        )
        Spacer(Modifier.height(8.dp))
        Surface(
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(8.dp),
            color = Color.White,
            border = BorderStroke(
                1.dp,
                if (isProblem) KimJbColorScheme.error else KimJbBorder,
            ),
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text(
                    item?.title?.ifBlank { title } ?: emptyText,
                    color = if (isProblem) KimJbColorScheme.error else KimJbInk,
                    fontSize = 15.sp,
                    fontWeight = FontWeight.SemiBold,
                )
                item?.body?.takeIf { it.isNotBlank() }?.let { body ->
                    Spacer(Modifier.height(6.dp))
                    Text(body, color = KimJbMuted, fontSize = 14.sp, lineHeight = 20.sp)
                }
            }
        }
    }
}

@Composable
private fun KimJbLink(
    title: String,
    description: String,
    onClick: () -> Unit,
) {
    Surface(
        onClick = onClick,
        modifier = Modifier
            .fillMaxWidth()
            .height(64.dp)
            .semantics { contentDescription = description },
        shape = RoundedCornerShape(8.dp),
        color = Color.White,
        contentColor = KimJbInk,
        border = BorderStroke(1.dp, KimJbBorder),
        tonalElevation = 0.dp,
        shadowElevation = 0.dp,
    ) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 18.dp),
            contentAlignment = androidx.compose.ui.Alignment.Center,
        ) {
            Text(
                title,
                color = KimJbInk,
                fontSize = 17.sp,
                fontWeight = FontWeight.SemiBold,
                textAlign = TextAlign.Center,
            )
        }
    }
}
