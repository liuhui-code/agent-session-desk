use std::collections::BTreeSet;
fn ids(p: &mut vt100_psmux::Parser) -> BTreeSet<u32> {
    let total = p.screen().scrollback_filled();
    let mut found = BTreeSet::new();
    for offset in 0..=total {
        p.screen_mut().set_scrollback(offset);
        for line in p.screen().contents().lines() {
            if let Some(id) = line
                .strip_prefix("history line ")
                .and_then(|s| s.trim().parse::<u32>().ok())
            {
                found.insert(id);
            }
        }
    }
    p.screen_mut().set_scrollback(0);
    found
}
fn main() {
    let mut p = vt100_psmux::Parser::new(32, 120, 2000);
    p.process("SCROLL_NEEDLE_中文_α\r\n".as_bytes());
    for i in 1..=40 {
        p.process(format!("history line {i}\r\n").as_bytes());
    }
    p.process(b"HISTORY_END\r\n> ");
    let before = ids(&mut p);
    println!(
        "before={before:?}; history={}",
        p.screen().scrollback_filled()
    );
    p.screen_mut().set_size(25, 94);
    let after = ids(&mut p);
    let lost: Vec<_> = before.difference(&after).collect();
    println!(
        "after={after:?}; history={}; lost={lost:?}",
        p.screen().scrollback_filled()
    );
    assert_eq!(
        before, after,
        "height shrink must retain numbered normal-screen output"
    );
}
