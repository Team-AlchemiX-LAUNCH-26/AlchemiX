interface PlanetInfo {
    id: string;
    codex: number;
  }
  
  interface ConversionHistoryProps {
    logs: unknown[];
    route: string[];
    planets: PlanetInfo[];
    originalPayload: string;
  }
  
  type HopRecord = Record<string, unknown>;
  
  const DIGITS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ";
  
  function convertToBase(value: number, base: number): string {
    if (base < 2 || base > DIGITS.length) {
      return String(value);
    }
  
    if (value === 0) {
      return "0";
    }
  
    let remaining = value;
    let output = "";
  
    while (remaining > 0) {
      output = DIGITS[remaining % base] + output;
      remaining = Math.floor(remaining / base);
    }
  
    return output;
  }
  
  function toRecord(value: unknown): HopRecord {
    if (
      typeof value === "object" &&
      value !== null &&
      !Array.isArray(value)
    ) {
      return value as HopRecord;
    }
  
    return {};
  }
  
  function readString(
    record: HopRecord,
    keys: string[]
  ): string | undefined {
    for (const key of keys) {
      const value = record[key];
  
      if (typeof value === "string" && value.length > 0) {
        return value;
      }
    }
  
    return undefined;
  }
  
  function readSequence(
    record: HopRecord,
    keys: string[]
  ): string[] | undefined {
    for (const key of keys) {
      const value = record[key];
  
      if (Array.isArray(value)) {
        return value.map((item) => String(item));
      }
  
      if (typeof value === "string" && value.trim()) {
        return value
          .replace(/^\[/, "")
          .replace(/\]$/, "")
          .split(/[,\s]+/)
          .filter(Boolean);
      }
    }
  
    return undefined;
  }
  
  function displayCharacter(character: string): string {
    if (character === " ") {
      return "␠";
    }
  
    if (character === "\n") {
      return "\\n";
    }
  
    if (character === "\t") {
      return "\\t";
    }
  
    return character;
  }
  
  export function ConversionHistory({
    logs,
    route,
    planets,
    originalPayload,
  }: ConversionHistoryProps) {
    const planetCodex = new Map(
      planets.map((planet) => [planet.id, planet.codex])
    );
  
    return (
      <div className="conversion-history">
        {route.map((planetID, routeIndex) => {
          const matchingLog =
            logs
              .map(toRecord)
              .find((log) => {
                const logPlanetID = readString(log, [
                  "planet_id",
                  "current_id",
                  "node_id",
                ]);
  
                return logPlanetID === planetID;
              }) ?? toRecord(logs[routeIndex]);
  
          const previousPlanetID =
            routeIndex > 0 ? route[routeIndex - 1] : undefined;
  
          const nextPlanetID =
            routeIndex < route.length - 1
              ? route[routeIndex + 1]
              : undefined;
  
          const currentCodex =
            planetCodex.get(planetID) ?? 10;
  
          const nextCodex = nextPlanetID
            ? planetCodex.get(nextPlanetID)
            : undefined;
  
          const decodedText =
            readString(matchingLog, [
              "decoded_payload",
              "ascii_payload",
              "local_payload",
              "decoded_text",
              "payload_text",
            ]) ?? originalPayload;
  
          const characters = Array.from(decodedText);
  
          const asciiValues = characters.map(
            (character) => character.codePointAt(0) ?? 0
          );
  
          const incomingSequence =
            readSequence(matchingLog, [
              "received_payload",
              "incoming_payload",
              "received_codes",
              "incoming_codes",
            ]) ??
            asciiValues.map((value) =>
              convertToBase(value, currentCodex)
            );
  
          const outgoingSequence = nextCodex
            ? readSequence(matchingLog, [
                "encoded_payload",
                "outgoing_payload",
                "translated_payload",
                "next_hop_payload",
                "outgoing_codes",
              ]) ??
              asciiValues.map((value) =>
                convertToBase(value, nextCodex)
              )
            : [];
  
          const binaryStream = readString(matchingLog, [
            "binary_stream",
            "binary_payload",
            "serialized_payload",
          ]);
  
          return (
            <article
              key={`${planetID}-${routeIndex}`}
              className="conversion-hop"
            >
              <div className="conversion-hop-header">
                <div>
                  <span className="conversion-hop-number">
                    Planet {routeIndex + 1} of {route.length}
                  </span>
  
                  <h4>{planetID}</h4>
                </div>
  
                <span className="conversion-route-label">
                  {previousPlanetID ?? "Origin"}
                  {" → "}
                  {planetID}
                  {" → "}
                  {nextPlanetID ?? "Delivered"}
                </span>
              </div>
  
              <div className="conversion-stage-flow">
                <span className="conversion-stage received">
                  {routeIndex === 0
                    ? `Local Base ${currentCodex}`
                    : `Received Base ${currentCodex}`}
                </span>
  
                <span className="conversion-stage-arrow">→</span>
  
                <span className="conversion-stage ascii">
                  ASCII
                </span>
  
                <span className="conversion-stage-arrow">→</span>
  
                {nextCodex ? (
                  <span className="conversion-stage outgoing">
                    Sent as Base {nextCodex}
                  </span>
                ) : (
                  <span className="conversion-stage delivered">
                    Message Delivered
                  </span>
                )}
              </div>
  
              <div className="conversion-message">
                <span>Decoded message</span>
                <strong>{decodedText}</strong>
              </div>
  
              <div className="conversion-table-scroll">
                <table className="conversion-table">
                  <thead>
                    <tr>
                      <th>Character</th>
  
                      <th>
                        {routeIndex === 0
                          ? `Local Base ${currentCodex}`
                          : `Received Base ${currentCodex}`}
                      </th>
  
                      <th>ASCII Decimal</th>
  
                      <th>
                        {nextCodex
                          ? `Next Base ${nextCodex}`
                          : "Final Text"}
                      </th>
                    </tr>
                  </thead>
  
                  <tbody>
                    {characters.map(
                      (character, characterIndex) => (
                        <tr
                          key={`${planetID}-${characterIndex}`}
                        >
                          <td>
                            <span className="character-cell">
                              {displayCharacter(character)}
                            </span>
                          </td>
  
                          <td>
                            <code>
                              {incomingSequence[
                                characterIndex
                              ] ??
                                convertToBase(
                                  asciiValues[characterIndex],
                                  currentCodex
                                )}
                            </code>
                          </td>
  
                          <td>
                            <code>
                              {asciiValues[characterIndex]}
                            </code>
                          </td>
  
                          <td>
                            <code>
                              {nextCodex
                                ? outgoingSequence[
                                    characterIndex
                                  ] ??
                                  convertToBase(
                                    asciiValues[
                                      characterIndex
                                    ],
                                    nextCodex
                                  )
                                : displayCharacter(character)}
                            </code>
                          </td>
                        </tr>
                      )
                    )}
                  </tbody>
                </table>
              </div>
  
              <div className="conversion-sequences">
                <div>
                  <span>
                    Base {currentCodex} input sequence
                  </span>
  
                  <code>
                    [{incomingSequence.join(", ")}]
                  </code>
                </div>
  
                <div>
                  <span>ASCII sequence</span>
  
                  <code>
                    [{asciiValues.join(", ")}]
                  </code>
                </div>
  
                {nextCodex && (
                  <div>
                    <span>
                      Base {nextCodex} output sequence
                    </span>
  
                    <code>
                      [{outgoingSequence.join(", ")}]
                    </code>
                  </div>
                )}
  
                {binaryStream && (
                  <div>
                    <span>Binary stream</span>
                    <code>{binaryStream}</code>
                  </div>
                )}
              </div>
            </article>
          );
        })}
      </div>
    );
  }