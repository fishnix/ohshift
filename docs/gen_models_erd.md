```mermaid
erDiagram
    incidents {
        text description 
        text export_url 
        uuid id PK 
        timestamp_with_time_zone last_updated 
        timestamp_with_time_zone resolved_at 
        character_varying resolved_by 
        character_varying severity 
        character_varying slack_channel_id 
        timestamp_with_time_zone started_at 
        character_varying started_by 
        character_varying status 
        text title 
    }

    timeline_events {
        character_varying event_type 
        uuid id PK 
        uuid incident_id FK 
        jsonb metadata 
        character_varying slack_message_ts 
        character_varying slack_user_id 
        timestamp_with_time_zone timestamp 
    }

    timeline_events }o--|| incidents : "incident_id"
```