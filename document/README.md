# Documentation

This directory contains UML diagrams for the video service.

- `videosequence/` contains sequence diagrams for the API endpoints.
- `entity_diagram.puml` shows the relationships between domain entities.

## Sequence diagrams

Existing diagrams describe the flow for `GetVideosByDate` and `GetVideosByKeyword`.
The following additional diagrams are provided:

- `get_video_by_id.puml` – sequence for fetching a single video by ID.
- `search_videos.puml` – sequence for searching videos using keyword or IDs.
- `get_videos_by_id.puml` – sequence for retrieving videos by multiple IDs.

## Entity diagram

`entity_diagram.puml` illustrates the Video entity and related structures such as
`Actress`, `Genre`, `Maker`, `Series`, `Director`, `Review`, and `SearchMetadata`.
