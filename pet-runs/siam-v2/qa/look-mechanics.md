# Siam look mechanics

Siam is a soft, grounded pixel-art kitten. The four paws and lower torso stay as the stable anchor while the head and upper chest make a restrained natural turn toward the target. The eyes lead the gaze: the original blue eye globes, dark mask, eyelids, pupils, rims, and highlights move together as one construction; no replacement googly eyes or pupil-only overlay is used. The ears follow the head with a small delay, and the curled tail stays attached to the rump while lagging slightly through the turn.

## Cardinal pose families

- `000` up: chin and eye globes tip upward, pupils sit high within the original eye construction, ears lift slightly; torso and paws remain planted.
- `090` screen-right: nose and face plane turn right, the right cheek and right-side ear become more prominent, the far cheek/ear is partly occluded; the tail remains attached and lags behind.
- `180` down: chin tucks toward the chest, eyelids lower and eye globes aim down, ears settle; lower body remains the anchor.
- `270` screen-left: nose and face plane turn left, the left cheek and left-side ear become more prominent, the far cheek/ear is partly occluded; the tail remains attached and lags behind.

## Interpolation and motion budget

Row 9 follows `000 -> 090 -> 180` through even 22.5-degree steps. Row 10 follows `180 -> 270 -> 000`, with `337.5` one step before the approved `000`. Each step changes the head/eye aim by a similar amount; no single neighbor may introduce a large body bend, scale change, prop shift, or silhouette jump. The lower-body anchor, baseline, fur palette, mask, blue eyes, and curled tail identity stay stable across all 16 cells.

## Existing expression/action coverage

The approved v1 standard rows are retained as Siam's expanded behavior set: calm blink/breathing idle, alternating right/left drag movement, raised-paw wave, airborne jump, slumped sleepy failure, expectant waiting, focused work, and attentive review. Look rows add directional eye/head expressions without changing the established pixel-art construction.
