# Vérifie si la touche F15 déclenche un effet observable sur ce PC.
# Stratégie : on note la fenêtre active avant/après l'envoi de plusieurs F15.
# Si F15 est inerte (cas attendu), rien ne change.

Add-Type @"
using System;
using System.Runtime.InteropServices;
using System.Text;
public static class Win {
    [DllImport("user32.dll")] public static extern IntPtr GetForegroundWindow();
    [DllImport("user32.dll")] public static extern int GetWindowText(IntPtr h, StringBuilder s, int n);
    [DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(IntPtr h, out uint pid);

    [StructLayout(LayoutKind.Sequential)]
    public struct INPUT { public uint type; public InputUnion U; }
    [StructLayout(LayoutKind.Explicit)]
    public struct InputUnion {
        [FieldOffset(0)] public KEYBDINPUT ki;
        [FieldOffset(0)] public MOUSEINPUT mi;
    }
    [StructLayout(LayoutKind.Sequential)]
    public struct KEYBDINPUT { public ushort wVk; public ushort wScan; public uint dwFlags; public uint time; public IntPtr extra; }
    [StructLayout(LayoutKind.Sequential)]
    public struct MOUSEINPUT { public int dx; public int dy; public uint data; public uint flags; public uint time; public IntPtr extra; }

    [DllImport("user32.dll")] public static extern uint SendInput(uint n, INPUT[] inputs, int size);

    public static void PressF15() {
        ushort VK_F15 = 0x7E;
        INPUT[] ins = new INPUT[2];
        ins[0].type = 1; ins[0].U.ki.wVk = VK_F15; ins[0].U.ki.dwFlags = 0;        // down
        ins[1].type = 1; ins[1].U.ki.wVk = VK_F15; ins[1].U.ki.dwFlags = 0x0002;   // up
        SendInput(2, ins, Marshal.SizeOf(typeof(INPUT)));
    }
}
"@

function Get-ActiveWindowInfo {
    $h = [Win]::GetForegroundWindow()
    $sb = New-Object System.Text.StringBuilder 512
    [void][Win]::GetWindowText($h, $sb, $sb.Capacity)
    $procId = 0
    [void][Win]::GetWindowThreadProcessId($h, [ref]$procId)
    $name = try { (Get-Process -Id $procId -ErrorAction Stop).ProcessName } catch { "?" }
    [PSCustomObject]@{ Handle = $h; Title = $sb.ToString(); Process = $name; Pid = $procId }
}

Write-Host "=== Test inertie de la touche F15 ===" -ForegroundColor Cyan
$before = Get-ActiveWindowInfo
Write-Host ("AVANT  : fenetre='{0}' process='{1}' (pid {2})" -f $before.Title, $before.Process, $before.Pid)

Write-Host "Envoi de 5 appuis F15..." -ForegroundColor Yellow
1..5 | ForEach-Object { [Win]::PressF15(); Start-Sleep -Milliseconds 200 }

$after = Get-ActiveWindowInfo
Write-Host ("APRES  : fenetre='{0}' process='{1}' (pid {2})" -f $after.Title, $after.Process, $after.Pid)
Write-Host ""

if ($before.Handle -eq $after.Handle) {
    Write-Host "RESULTAT : la fenetre active n'a PAS change. F15 semble inerte sur ce PC." -ForegroundColor Green
} else {
    Write-Host "RESULTAT : la fenetre active A CHANGE ! Quelque chose reagit a F15 :" -ForegroundColor Red
    Write-Host ("  -> nouvelle fenetre = '{0}' ({1})" -f $after.Title, $after.Process)
}
