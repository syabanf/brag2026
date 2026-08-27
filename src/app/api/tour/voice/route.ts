import { TOUR_STEPS } from "@/lib/tour-steps";

const ELEVENLABS_MODEL = "eleven_multilingual_v2";

// Narration is identical on every playback, so each clip is synthesised once
// per server lifetime and replayed from memory afterwards.
const audioCache = new Map<string, ArrayBuffer>();

export async function GET(request: Request) {
  const stepId = new URL(request.url).searchParams.get("step");
  const step = TOUR_STEPS.find((s) => s.id === stepId);

  if (!step) {
    return Response.json({ error: "Unknown tour step" }, { status: 404 });
  }

  const apiKey = process.env.ELEVENLABS_API_KEY;
  const voiceId = process.env.ELEVENLABS_VOICE_ID;

  // Voice-over is optional: without credentials the tour runs silently.
  if (!apiKey || !voiceId) {
    return new Response(null, { status: 204 });
  }

  const cached = audioCache.get(step.id);
  if (cached) {
    return new Response(cached, {
      headers: { "content-type": "audio/mpeg", "x-cache": "hit" }
    });
  }

  const upstream = await fetch(
    `https://api.elevenlabs.io/v1/text-to-speech/${voiceId}`,
    {
      method: "POST",
      headers: {
        "xi-api-key": apiKey,
        "content-type": "application/json"
      },
      body: JSON.stringify({
        text: step.voice,
        model_id: ELEVENLABS_MODEL,
        voice_settings: { stability: 0.5, similarity_boost: 0.75 }
      })
    }
  );

  if (!upstream.ok) {
    const detail = await upstream.text();
    console.error("ElevenLabs TTS failed:", upstream.status, detail.slice(0, 300));

    if (upstream.status === 402) {
      console.error(
        "Hint: free ElevenLabs plans cannot use library/professional voices via the API. " +
          "Set ELEVENLABS_VOICE_ID to a premade voice, or upgrade the plan. " +
          "The tour falls back to browser speech synthesis in the meantime."
      );
    }

    // The client treats any non-audio response as \"use the browser voice\".
    return Response.json({ error: "Text-to-speech unavailable" }, { status: 502 });
  }

  const audio = await upstream.arrayBuffer();
  audioCache.set(step.id, audio);

  return new Response(audio, {
    headers: { "content-type": "audio/mpeg", "x-cache": "miss" }
  });
}
