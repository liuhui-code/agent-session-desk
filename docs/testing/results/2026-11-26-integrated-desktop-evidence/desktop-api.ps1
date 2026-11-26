Add-Type -AssemblyName UIAutomationClient
Add-Type -AssemblyName UIAutomationTypes
Add-Type @'
using System;using System.Text;using System.Runtime.InteropServices;using System.Collections.Generic;
public class DeskFlow {
 public delegate int PropProc(IntPtr h,IntPtr name,IntPtr data,IntPtr extra);
 [DllImport("user32.dll",CharSet=CharSet.Unicode)]public static extern int EnumPropsEx(IntPtr h,PropProc callback,IntPtr extra);
 public static int OwnerBindings(IntPtr h){int count=0;EnumPropsEx(h,(w,n,d,x)=>{if(n.ToInt64()>65535){string s=Marshal.PtrToStringUni(n);if(s.StartsWith("ASD.Attention.Owner.",StringComparison.Ordinal))count++;}return 1;},IntPtr.Zero);return count;}
 [DllImport("user32.dll",CharSet=CharSet.Unicode)]public static extern IntPtr FindWindowEx(IntPtr parent,IntPtr after,string klass,string name);
 [DllImport("user32.dll",CharSet=CharSet.Unicode)]public static extern IntPtr SendMessageTimeout(IntPtr h,uint m,UIntPtr w,StringBuilder b,uint flags,uint timeout,out UIntPtr result);
 public static bool EditContains(IntPtr parent,int pid,string marker){IntPtr edit=FindWindowEx(parent,IntPtr.Zero,"Edit",null);uint owner;GetWindowThreadProcessId(edit,out owner);if(edit==IntPtr.Zero||owner!=(uint)pid)return false;var b=new StringBuilder(2048);UIntPtr result;return SendMessageTimeout(edit,0x000D,(UIntPtr)2048,b,2,1000,out result)!=IntPtr.Zero&&b.ToString().Contains(marker);}
 [DllImport("user32.dll")]public static extern IntPtr SetThreadDpiAwarenessContext(IntPtr context);
 [DllImport("user32.dll")]public static extern bool GetLastInputInfo(ref LastInput info);
 [StructLayout(LayoutKind.Sequential)]public struct LastInput{public uint Size,Time;}
 public static uint IdleMilliseconds(){var x=new LastInput();x.Size=(uint)Marshal.SizeOf(typeof(LastInput));if(!GetLastInputInfo(ref x))return 0;return unchecked((uint)Environment.TickCount-x.Time);}
 public static IntPtr[] HostWindows(string path){var list=new List<IntPtr>();EnumWindows((h,x)=>{if(!IsWindowVisible(h))return true;uint id;GetWindowThreadProcessId(h,out id);try{using(var p=System.Diagnostics.Process.GetProcessById((int)id)){if(String.Equals(p.MainModule.FileName,path,StringComparison.OrdinalIgnoreCase))list.Add(h);}}catch{}return true;},IntPtr.Zero);return list.ToArray();}
 [DllImport("shlwapi.dll",CharSet=CharSet.Unicode)]public static extern int AssocQueryString(uint flags,uint kind,string assoc,string extra,StringBuilder output,ref uint length);
 public static string AssociationStamp(){var b=new StringBuilder();foreach(uint kind in new uint[]{20,2}){uint length=2048;var value=new StringBuilder(2048);int hr=AssocQueryString(0x400,kind,".md","open",value,ref length);b.Append(hr).Append(':').Append(value).Append(';');}using(var sha=System.Security.Cryptography.SHA256.Create()){return Convert.ToBase64String(sha.ComputeHash(Encoding.Unicode.GetBytes(b.ToString())));}}
 public delegate bool EnumProc(IntPtr h,IntPtr p);
 [DllImport("user32.dll")]public static extern bool EnumWindows(EnumProc e,IntPtr p);
 [DllImport("user32.dll")]public static extern uint GetWindowThreadProcessId(IntPtr h,out uint p);
 [DllImport("user32.dll",CharSet=CharSet.Unicode)]public static extern int GetWindowText(IntPtr h,StringBuilder b,int n);
 [DllImport("user32.dll")]public static extern bool IsWindowVisible(IntPtr h);
 [DllImport("user32.dll")]public static extern IntPtr GetForegroundWindow();
 [DllImport("user32.dll")]public static extern bool SetForegroundWindow(IntPtr h);
 [DllImport("user32.dll")]public static extern short GetAsyncKeyState(int k);
 [DllImport("user32.dll")]public static extern bool GetWindowRect(IntPtr h,out R r);
 [DllImport("user32.dll")]public static extern bool GetCursorPos(out P p);
 [DllImport("user32.dll")]public static extern int GetSystemMetrics(int k);
 [DllImport("user32.dll")]public static extern uint SendInput(uint n,I[] a,int size);
 [DllImport("user32.dll")]public static extern IntPtr OpenInputDesktop(uint flags,bool inherit,uint access);
 [DllImport("user32.dll")]public static extern bool CloseDesktop(IntPtr h);
 [DllImport("user32.dll")]public static extern IntPtr SendMessage(IntPtr h,uint m,UIntPtr w,IntPtr l);
 [StructLayout(LayoutKind.Sequential)]public struct P{public int X,Y;}
 [StructLayout(LayoutKind.Sequential)]public struct R{public int Left,Top,Right,Bottom;}
 [StructLayout(LayoutKind.Sequential)]public struct M{public int x,y;public uint data,flags,time;public UIntPtr extra;}
 [StructLayout(LayoutKind.Sequential)]public struct K{public ushort key,scan;public uint flags,time;public UIntPtr extra;}
 [StructLayout(LayoutKind.Explicit)]public struct D{[FieldOffset(0)]public M m;[FieldOffset(0)]public K k;}
 [StructLayout(LayoutKind.Sequential)]public struct I{public uint type;public D d;}
 public static string Title(IntPtr h){var b=new StringBuilder(512);GetWindowText(h,b,512);return b.ToString();}
 public static IntPtr[] Find(string prefix,int pid=0){var result=new List<IntPtr>();EnumWindows((h,p)=>{uint owner;GetWindowThreadProcessId(h,out owner);if((pid==0||owner==(uint)pid)&&Title(h).StartsWith(prefix,StringComparison.Ordinal)&&IsWindowVisible(h))result.Add(h);return true;},IntPtr.Zero);return result.ToArray();}
 public static uint Keys(ushort[] keys,uint[] flags){var a=new I[keys.Length];for(int n=0;n<a.Length;n++){a[n].type=1;a[n].d.k.key=keys[n];a[n].d.k.flags=flags[n];}return SendInput((uint)a.Length,a,Marshal.SizeOf(typeof(I)));}
 public static uint Key(ushort key){return Keys(new ushort[]{key,key},new uint[]{0,2});}
 public static uint Chord(ushort mod,ushort key){return Keys(new ushort[]{mod,key,key,mod},new uint[]{0,0,2,2});}
 public static uint Text(string text){var a=new I[text.Length*2];for(int n=0;n<text.Length;n++){a[2*n].type=1;a[2*n].d.k.scan=text[n];a[2*n].d.k.flags=4;a[2*n+1]=a[2*n];a[2*n+1].d.k.flags=6;}return SendInput((uint)a.Length,a,Marshal.SizeOf(typeof(I)));}
 public static uint Move(int x,int y){int l=GetSystemMetrics(76),t=GetSystemMetrics(77),w=GetSystemMetrics(78),h=GetSystemMetrics(79);var a=new I[1];a[0].d.m.x=(int)Math.Round((x-l)*65535.0/(w-1));a[0].d.m.y=(int)Math.Round((y-t)*65535.0/(h-1));a[0].d.m.flags=0xC001;return SendInput(1,a,Marshal.SizeOf(typeof(I)));}
 public static uint Click(){var a=new I[2];a[0].d.m.flags=2;a[1].d.m.flags=4;return SendInput(2,a,Marshal.SizeOf(typeof(I)));}
}
'@
