# Deterministic Monthly OpenSpec Archive

The monthly OpenSpec archive workflow currently asks an AI tool to perform the base archive step. The failed run `26166669462` showed that the AI process can exit successfully while leaving completed active changes unarchived, so the workflow fails later in the assertion step without producing a partial archive PR.

This change moves the base archive step into a deterministic GitHub Action that scans completed active OpenSpec changes and runs `openspec archive -y <change>` directly. AI tools remain available for archive consolidation after deterministic archiving produces changes. The workflow should still surface unarchived completed changes with clear failure details, but it should be able to create or update the archive PR for successful archive results before failing on the remaining blockers.

The same change fixes the known `remove-sqlite-support` archive blocker by aligning its OpenSpec delta with the current `cluster-coordination-config` baseline, and updates artifact upload actions away from the Node 20 runtime generation.
