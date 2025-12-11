from __future__ import annotations

import tkinter as tk
from tkinter import messagebox
from datetime import datetime, time, timedelta, timezone
from typing import List

from .storage import Entry, append_entry, find_open, read_entries, write_entries


def local_now() -> datetime:
    return datetime.now().astimezone().replace(microsecond=0)


def to_utc(value: datetime) -> datetime:
    if value.tzinfo is None:
        value = value.replace(tzinfo=timezone.utc)
    return value.astimezone(timezone.utc)


def format_duration(delta: timedelta) -> str:
    total_seconds = int(delta.total_seconds())
    hours, remainder = divmod(total_seconds, 3600)
    minutes, _ = divmod(remainder, 60)
    return f"{hours}h{minutes:02d}m"


def format_hhmm(delta: timedelta) -> str:
    total_minutes = int(delta.total_seconds() // 60)
    hours, minutes = divmod(total_minutes, 60)
    return f"{hours:02d}:{minutes:02d}"


def format_entry_line(entry: Entry, now: datetime, tzinfo) -> str:
    start_local = entry.start.astimezone(tzinfo)
    end_local = (entry.end or now).astimezone(tzinfo)
    duration = entry.duration(now)
    end_label = end_local.strftime("%H:%M") if entry.end else "…"
    return (
        f"{start_local:%H:%M} - {end_label}  "
        f"{format_duration(duration):>6}  {entry.text}"
    )


class TimeLogApp:
    def __init__(self, root: tk.Tk) -> None:
        self.root = root
        self.root.title("pytimelog")
        self.entries: List[Entry] = []

        self.status_label = tk.Label(root, text="Loading…", anchor="w", font=("Helvetica", 12))
        self.status_label.pack(fill="x", padx=8, pady=(8, 4))

        self.summary_label = tk.Label(root, text="", anchor="w", font=("Helvetica", 11))
        self.summary_label.pack(fill="x", padx=8, pady=(0, 6))

        form = tk.Frame(root)
        form.pack(fill="x", padx=8, pady=4)
        tk.Label(form, text="What are you doing?").pack(anchor="w")
        self.entry_text = tk.Entry(form)
        self.entry_text.pack(fill="x", pady=4)

        buttons = tk.Frame(form)
        buttons.pack(fill="x", pady=(0, 4))
        self.start_button = tk.Button(buttons, text="Start", command=self.start_entry)
        self.stop_button = tk.Button(buttons, text="Stop", command=self.stop_entry)
        self.start_button.pack(side="left")
        self.stop_button.pack(side="left", padx=(6, 0))

        tk.Label(root, text="Today's entries").pack(anchor="w", padx=8, pady=(6, 0))
        self.listbox = tk.Listbox(root, height=12, font=("Courier", 11))
        self.listbox.pack(fill="both", expand=True, padx=8, pady=(2, 8))

        self.refresh()

    def refresh(self) -> None:
        self.entries = read_entries()
        now = to_utc(local_now())
        self.render_status(now)
        self.render_today(now)
        self.root.after(15000, self.refresh)

    def render_status(self, now: datetime) -> None:
        idx = find_open(self.entries)
        if idx is None:
            self.status_label.config(text="Status: idle")
        else:
            entry = self.entries[idx]
            elapsed = entry.duration(now)
            self.status_label.config(
                text=f"Status: running '{entry.text}' ({format_duration(elapsed)} elapsed)"
            )
        self.render_summary(now)

    def render_today(self, now: datetime) -> None:
        tzinfo = now.astimezone().tzinfo
        start_local, end_local = self.today_window(now)

        self.listbox.delete(0, tk.END)
        for entry in self.entries:
            entry_end = entry.end or now
            if entry.start >= end_local or entry_end <= start_local:
                continue
            line = format_entry_line(entry, now, tzinfo)
            self.listbox.insert(tk.END, line)

    def today_window(self, now: datetime) -> tuple[datetime, datetime]:
        tzinfo = now.astimezone().tzinfo
        today = now.astimezone().date()
        start_local = datetime.combine(today, time.min, tzinfo=tzinfo).astimezone(timezone.utc)
        end_local = datetime.combine(today + timedelta(days=1), time.min, tzinfo=tzinfo).astimezone(timezone.utc)
        return start_local, end_local

    def render_summary(self, now: datetime) -> None:
        start_local, end_local = self.today_window(now)
        total = timedelta(0)
        for entry in self.entries:
            entry_end = entry.end or now
            latest_start = max(entry.start, start_local)
            earliest_end = min(entry_end, end_local)
            if earliest_end <= latest_start:
                continue
            total += earliest_end - latest_start

        target = timedelta(hours=8)
        remaining = target - total
        if remaining < timedelta(0):
            remaining = timedelta(0)
        self.summary_label.config(
            text=f"Today: {format_hhmm(total)} worked, {format_hhmm(remaining)} remaining to 08:00"
        )

    def start_entry(self) -> None:
        text = self.entry_text.get().strip()
        if not text:
            messagebox.showwarning("pytimelog", "Please enter a description.")
            return
        entries = read_entries()
        if find_open(entries) is not None:
            messagebox.showwarning("pytimelog", "An entry is already running.")
            return
        start = to_utc(local_now())
        append_entry(Entry(start=start, end=None, text=text))
        self.entry_text.delete(0, tk.END)
        self.refresh()

    def stop_entry(self) -> None:
        entries = read_entries()
        idx = find_open(entries)
        if idx is None:
            messagebox.showinfo("pytimelog", "No active entry to stop.")
            return
        now = to_utc(local_now())
        open_entry = entries[idx]
        if now <= open_entry.start:
            now = open_entry.start + timedelta(minutes=1)
        entries[idx] = Entry(start=open_entry.start, end=now, text=open_entry.text)
        write_entries(entries)
        self.refresh()


def launch_ui() -> None:
    root = tk.Tk()
    app = TimeLogApp(root)
    root.mainloop()


if __name__ == "__main__":
    launch_ui()
